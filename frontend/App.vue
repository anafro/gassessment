<script setup lang="ts">
    import { get, set } from '@vueuse/core';
    import { computed, onMounted, onUnmounted, ref } from 'vue';
    import {
        recoverPending,
        pendingStorageKey,
        type BetRequest,
    } from './recovery';
    import { money } from './lib/format';
    /** Server line snapshot; odds are integer hundredths in P1/draw/P2 order. */
    type Match = {
        id: number;
        name: string;
        odds: number[];
    };
    /** Frozen price chosen by the player, independent of subsequent live updates. */
    type Selection = {
        match: number;
        outcome: number;
        odds: number;
    };
    const matches = ref<Match[]>([]);
    const balance = ref<number>();
    const connected = ref(false);
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
    let events: EventSource | undefined;
    onUnmounted(() => events?.close());
    // Wallet loads once; EventSource reconnects automatically after transport errors.
    onMounted(async () => {
        try {
            const response = await fetch('/api/wallet');
            if (!response.ok) throw new Error();
            set(balance, (await response.json()).balance);
        } catch {
            set(message, 'Не удалось загрузить баланс');
        }
        events = new EventSource('/api/events');
        events.onmessage = (e) => {
            set(matches, JSON.parse(e.data));
            set(connected, true);
        };
        events.onerror = () => set(connected, false);
    });
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
</script>
<template>
    <main class="mx-auto my-9 max-w-[950px] p-5">
        <header class="flex flex-wrap items-center justify-between gap-4">
            <h1 class="my-5 text-[28px] font-bold">Мини-спортсбук</h1>
            <strong class="font-bold"
                >Баланс:
                {{
                    balance === undefined ? '—' : money(balance) + ' ₽'
                }}</strong
            >
        </header>
        <p
            class="my-4"
            :class="connected ? 'text-[#187747]' : 'text-[#b74121]'"
        >
            {{
                connected
                    ? '● Линия обновляется'
                    : '● Связь потеряна, переподключаемся…'
            }}
        </p>
        <section
            class="grid grid-cols-[repeat(auto-fit,minmax(min(260px,100%),1fr))] gap-3.5"
        >
            <article
                v-for="m in matches"
                :key="m.id"
                class="rounded-[10px] bg-white p-5"
            >
                <h3 class="mb-4 font-bold">{{ m.name }}</h3>
                <div class="flex gap-2">
                    <button
                        v-for="(odd, i) in m.odds"
                        :key="i"
                        type="button"
                        class="flex-1 cursor-pointer rounded-md border border-slate-300 p-3 disabled:cursor-default disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                        :disabled="
                            storageBlocked || !connected || busy || !!pending
                        "
                        :class="
                            selected?.match === m.id && selected?.outcome === i
                                ? 'bg-[#215bd7] text-white'
                                : 'bg-[#e8eef7] text-[#192637]'
                        "
                        @click="
                            selected = { match: m.id, outcome: i, odds: odd }
                        "
                    >
                        {{ ['П1', 'Х', 'П2'][i] }}
                        <b class="ml-2.5 font-bold">{{ money(odd) }}</b>
                    </button>
                </div>
            </article>
        </section>
        <aside class="mt-5 rounded-[10px] bg-white p-5">
            <h2 class="my-5 text-2xl font-bold">Купон</h2>
            <p v-if="selected" class="my-4">
                {{ selectedMatchName }} ·
                {{ ['П1', 'Х', 'П2'][selected.outcome] }} @
                {{ money(selected.odds) }}
            </p>
            <p v-else class="my-4">Выберите исход</p>
            <label
                >Сумма, ₽
                <input
                    v-model="amount"
                    inputmode="decimal"
                    class="mx-3 w-30 rounded-md border border-slate-300 p-3"
                    :disabled="storageBlocked || busy || !!pending" /></label
            ><button
                type="button"
                class="cursor-pointer rounded-md border border-slate-300 bg-[#215bd7] p-3 text-white disabled:cursor-default disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                :disabled="
                    storageBlocked ||
                    busy ||
                    (!pending &&
                        (!selected || !connected || balance === undefined))
                "
                @click="submit"
            >
                {{
                    busy
                        ? 'Отправка…'
                        : pending
                          ? 'Повторить запрос'
                          : 'Поставить'
                }}
            </button>
            <p role="status" class="my-4">{{ message }}</p>
        </aside>
    </main>
</template>
