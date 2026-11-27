import { ru } from '../locales/ru';
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
        set(message, ru.messages.recoveryBlocked);
    } else if (recovery.pending) {
        set(selected, recovery.pending);
        set(message, ru.messages.pendingRecovered);
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

    // Keep the accepted selection frozen until the player confirms a displayed price.
    const offeredOdds = computed(() => {
        const selection = get(selected);
        if (!selection || get(pending)) return undefined;
        const odds = get(matches).find((match) => match.id === selection.match)
            ?.odds[selection.outcome];
        return odds !== undefined && odds !== selection.odds ? odds : undefined;
    });
    function confirmOdds(displayedOdds: number) {
        if (
            get(editDisabled) ||
            !get(connected) ||
            get(offeredOdds) !== displayedOdds
        )
            return;
        const selection = get(selected);
        if (selection) set(selected, { ...selection, odds: displayedOdds });
    }

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
            get(offeredOdds) !== undefined ||
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
        if (get(submitDisabled)) return;
        if (!get(pending)) {
            const selection = get(selected);
            if (!selection || !/^\d{1,7}(\.\d{1,2})?$/.test(get(amount))) {
                set(message, ru.messages.invalidAmount);
                return;
            }
            const [rub, kop = ''] = get(amount).split('.');
            const cents = Number(rub) * 100 + Number(kop.padEnd(2, '0'));
            if (cents <= 0) {
                set(message, ru.messages.positiveAmount);
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
            set(message, ru.messages.persistenceFailed);
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
                    typeof data === 'string' ? data : ru.messages.serverError,
                );
            }
            if (
                data.accepted !== true ||
                !Number.isSafeInteger(data.balance) ||
                data.balance < 0
            ) {
                throw new Error(ru.messages.invalidResponse);
            }
            set(balance, data.balance);
            clearPending();
            set(selected, undefined);
            set(message, ru.messages.accepted);
        } catch (e) {
            set(
                message,
                (e instanceof Error ? e.message : ru.messages.networkError) +
                    (get(pending) ? ru.messages.retryHint : ''),
            );
        } finally {
            set(busy, false);
        }
    }

    return {
        amount,
        selected,
        selectedMatchName,
        offeredOdds,
        confirmOdds,
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
