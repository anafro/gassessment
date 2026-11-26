import { get, set } from '@vueuse/core';
import { computed, ref, type Ref } from 'vue';
import {
    recoverPending,
    pendingStorageKey,
    type BetRequest,
} from '../recovery';
import type { Match, Selection } from '../types';

type BetSlipOptions = {
    matches: Ref<Match[]>;
    balance: Ref<number | undefined>;
    connected: Ref<boolean>;
};

/** Owns the bet slip and durable same-key retries; the server owns acceptance. */
export function useBetSlip({ matches, balance, connected }: BetSlipOptions) {
    const message = ref('');
    const amount = ref('100.00');
    const busy = ref(false);
    const selected = ref<Selection>();
    // Access localStorage inside the callback so even a denied getter is caught.
    const recovery = recoverPending(
        () => localStorage.getItem(pendingStorageKey),
        () => localStorage.removeItem(pendingStorageKey),
    );
    const pending = ref<BetRequest | undefined>(recovery.pending);
    const storageBlocked = ref(recovery.blocked);
    if (recovery.blocked) {
        set(
            message,
            'Не удалось восстановить запрос. Новые ставки заблокированы: проверьте хранилище браузера.',
        );
    } else if (recovery.pending) {
        set(selected, recovery.pending);
        set(message, 'Есть незавершённый запрос. Повторите его для проверки.');
    }
    /** Remove durable state first; a storage failure keeps the retry in memory. */
    function clearPending() {
        localStorage.removeItem(pendingStorageKey);
        set(pending, undefined);
    }
    const selectedMatchName = computed(
        () =>
            get(matches).find((match) => match.id === get(selected)?.match)
                ?.name,
    );

    function selectOutcome(selection: Selection) {
        if (get(selectionDisabled)) return;
        set(selected, selection);
    }
    const editDisabled = computed(
        () => get(storageBlocked) || get(busy) || !!get(pending),
    );
    const selectionDisabled = computed(
        () => get(editDisabled) || !get(connected),
    );
    const submitDisabled = computed(
        () =>
            get(storageBlocked) ||
            get(busy) ||
            (!get(pending) &&
                (!get(selected) ||
                    !get(connected) ||
                    get(balance) === undefined)),
    );
    /**
     * Persist before sending and reuse the payload on uncertain results.
     * Clear it only on validated success or a definitive application rejection.
     * busy prevents overlapping submissions within this mounted component.
     */
    async function submit() {
        if (get(busy) || get(storageBlocked)) return;
        if (!get(pending)) {
            const selection = get(selected);
            if (!selection || !/^\d{1,7}(\.\d{1,2})?$/.test(get(amount))) {
                set(message, 'Введите сумму с точностью до копеек');
                return;
            }
            const [rub, kop = ''] = get(amount).split('.');
            const cents = Number(rub) * 100 + Number(kop.padEnd(2, '0'));
            if (cents <= 0) {
                set(message, 'Сумма должна быть положительной');
                return;
            }
            set(pending, {
                ...selection,
                key: crypto.randomUUID(),
                amount: String(cents),
            });
        }
        try {
            localStorage.setItem(
                pendingStorageKey,
                JSON.stringify(get(pending)),
            );
        } catch {
            set(message, 'Не удалось сохранить запрос. Ставка не отправлена.');
            return;
        }
        set(busy, true);
        try {
            const response = await fetch('/api/bets', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(get(pending)),
            });
            const data = await response.json();
            if (!response.ok) {
                if (
                    [400, 409, 422].includes(response.status) &&
                    typeof data === 'string'
                )
                    clearPending();
                throw new Error(
                    typeof data === 'string' ? data : 'Ошибка сервера',
                );
            }
            if (
                data.accepted !== true ||
                !Number.isSafeInteger(data.balance) ||
                data.balance < 0
            ) {
                throw new Error('Некорректный ответ сервера');
            }
            set(balance, data.balance);
            clearPending();
            set(selected, undefined);
            set(message, 'Ставка принята');
        } catch (e) {
            set(
                message,
                (e instanceof Error ? e.message : 'Ошибка сети') +
                    (get(pending)
                        ? ' Повторите: запрос сохранён, двойного списания не будет.'
                        : ''),
            );
        } finally {
            set(busy, false);
        }
    }

    return {
        amount,
        selected,
        selectedMatchName,
        pending,
        busy,
        message,
        editDisabled,
        selectionDisabled,
        submitDisabled,
        selectOutcome,
        submit,
    };
}
