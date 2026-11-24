<script setup lang="ts">
    import { computed, onMounted, onUnmounted, ref } from 'vue';
    /** Server line snapshot; odds are integer hundredths in P1/draw/P2 order. */
    type Match = { id: number; name: string; odds: number[] };
    /** Frozen price chosen by the player, independent of subsequent live updates. */
    type Selection = { match: number; outcome: number; odds: number };
    /** Payload preserved unchanged until the server gives a definitive result. */
    type BetRequest = Selection & { key: string; amount: string };
    const matches = ref<Match[]>([]);
    const balance = ref<number>();
    const connected = ref(false);
    const message = ref('');
    const amount = ref('100.00');
    const busy = ref(false);
    const selected = ref<Selection>();
    const pending = ref<BetRequest>();
    const storageBlocked = ref(false);
    const pendingStorageKey = 'sportsbook.pending.v1';
    // Recovery is best effort; saved JSON currently has no runtime schema check.
    try {
        const saved = localStorage.getItem(pendingStorageKey);
        if (saved) {
            pending.value = JSON.parse(saved);
            selected.value = pending.value;
            message.value =
                'Есть незавершённый запрос. Повторите его для проверки.';
        }
    } catch {
        message.value =
            'Хранилище недоступно. Отправка новых ставок заблокирована.';
    }
    /** Remove durable state first; a storage failure keeps the retry in memory. */
    function clearPending() {
        localStorage.removeItem(pendingStorageKey);
        pending.value = undefined;
    }
    const selectedMatchName = computed(
        () =>
            matches.value.find((match) => match.id === selected.value?.match)
                ?.name,
    );
    let events: EventSource | undefined;
    onUnmounted(() => events?.close());
    /** Display minor units or odds hundredths; never use this for balance math. */
    const money = (n: number) => (n / 100).toFixed(2);
    // Wallet loads once; EventSource reconnects automatically after transport errors.
    onMounted(async () => {
        try {
            const response = await fetch('/api/wallet');
            if (!response.ok) throw new Error();
            balance.value = (await response.json()).balance;
        } catch {
            message.value = 'Не удалось загрузить баланс';
        }
        events = new EventSource('/api/events');
        events.onmessage = (e) => {
            matches.value = JSON.parse(e.data);
            connected.value = true;
        };
        events.onerror = () => (connected.value = false);
    });
    /**
     * Persist before sending and reuse the payload on uncertain results.
     * Clear it only on validated success or a definitive application rejection.
     * busy prevents overlapping submissions within this mounted component.
     */
    async function submit() {
        if (busy.value || storageBlocked.value) return;
        if (!pending.value) {
            if (
                !selected.value ||
                !/^\d{1,7}(\.\d{1,2})?$/.test(amount.value)
            ) {
                message.value = 'Введите сумму с точностью до копеек';
                return;
            }
            const [rub, kop = ''] = amount.value.split('.');
            const cents = Number(rub) * 100 + Number(kop.padEnd(2, '0'));
            if (cents <= 0) {
                message.value = 'Сумма должна быть положительной';
                return;
            }
            pending.value = {
                ...selected.value,
                key: crypto.randomUUID(),
                amount: String(cents),
            };
        }
        try {
            localStorage.setItem(
                pendingStorageKey,
                JSON.stringify(pending.value),
            );
        } catch {
            message.value =
                'Не удалось сохранить запрос. Ставка не отправлена.';
            return;
        }
        busy.value = true;
        try {
            const response = await fetch('/api/bets', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(pending.value),
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
            balance.value = data.balance;
            clearPending();
            selected.value = undefined;
            message.value = 'Ставка принята';
        } catch (e) {
            message.value =
                (e instanceof Error ? e.message : 'Ошибка сети') +
                (pending.value
                    ? ' Повторите: запрос сохранён, двойного списания не будет.'
                    : '');
        } finally {
            busy.value = false;
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
                        :disabled="!connected || busy || !!pending"
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
                    :disabled="busy || !!pending" /></label
            ><button
                type="button"
                class="cursor-pointer rounded-md border border-slate-300 bg-[#215bd7] p-3 text-white disabled:cursor-default disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                :disabled="
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
