<script setup lang="ts">
    import { ref, onMounted } from 'vue';
    type Match = { id: number; name: string; odds: number[] };
    const matches = ref<Match[]>([]),
        balance = ref<number>(),
        connected = ref(false),
        message = ref(''),
        amount = ref('100.00'),
        busy = ref(false);
    const selected = ref<{ match: number; outcome: number; odds: number }>();
    let pending:
        | {
              key: string;
              match: number;
              outcome: number;
              odds: number;
              amount: string;
          }
        | undefined;
    const money = (n: number) => (n / 100).toFixed(2);
    onMounted(async () => {
        try {
            const r = await fetch('/api/wallet');
            if (!r.ok) throw Error();
            balance.value = (await r.json()).balance;
        } catch {
            message.value = 'Не удалось загрузить баланс';
        }
        const events = new EventSource('/api/events');
        events.onmessage = (e) => {
            matches.value = JSON.parse(e.data);
            connected.value = true;
        };
        events.onerror = () => (connected.value = false);
    });
    async function submit() {
        if (busy.value) return;
        if (!pending) {
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
            pending = {
                ...selected.value,
                key: crypto.randomUUID(),
                amount: String(cents),
            };
        }
        busy.value = true;
        try {
            const r = await fetch('/api/bets', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(pending),
            });
            const data = await r.json();
            if (!r.ok) {
                if (r.status < 500) pending = undefined;
                throw Error(typeof data === 'string' ? data : 'Ошибка сервера');
            }
            balance.value = data.balance;
            pending = undefined;
            selected.value = undefined;
            message.value = 'Ставка принята';
        } catch (e) {
            message.value =
                (e instanceof Error ? e.message : 'Ошибка сети') +
                (pending
                    ? ' Повторите: запрос сохранён, двойного списания не будет.'
                    : '');
        } finally {
            busy.value = false;
        }
    }
</script>
<template>
    <main>
        <header>
            <h1>Мини-спортсбук</h1>
            <strong
                >Баланс:
                {{
                    balance === undefined ? '—' : money(balance) + ' ₽'
                }}</strong
            >
        </header>
        <p :class="connected ? 'online' : 'offline'">
            {{
                connected
                    ? '● Линия обновляется'
                    : '● Связь потеряна, переподключаемся…'
            }}
        </p>
        <section>
            <article v-for="m in matches" :key="m.id">
                <h3>{{ m.name }}</h3>
                <div class="odds">
                    <button
                        v-for="(odd, i) in m.odds"
                        :key="i"
                        :disabled="!connected || busy || !!pending"
                        :class="{
                            chosen:
                                selected?.match === m.id &&
                                selected?.outcome === i,
                        }"
                        @click="
                            selected = { match: m.id, outcome: i, odds: odd }
                        "
                    >
                        {{ ['П1', 'Х', 'П2'][i] }} <b>{{ money(odd) }}</b>
                    </button>
                </div>
            </article>
        </section>
        <aside>
            <h2>Купон</h2>
            <p v-if="selected">
                {{ matches.find((m) => m.id === selected?.match)?.name }} ·
                {{ ['П1', 'Х', 'П2'][selected.outcome] }} @
                {{ money(selected.odds) }}
            </p>
            <p v-else>Выберите исход</p>
            <label
                >Сумма, ₽
                <input
                    v-model="amount"
                    inputmode="decimal"
                    :disabled="busy || !!pending" /></label
            ><button
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
            <p role="status">{{ message }}</p>
        </aside>
    </main>
</template>
<style>
    body {
        margin: 0;
        background: #f3f5f8;
        color: #192637;
        font: 16px system-ui;
    }

    main {
        max-width: 950px;
        margin: 35px auto;
        padding: 20px;
    }

    header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    h1 {
        font-size: 28px;
    }

    section {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
        gap: 14px;
    }

    article,
    aside {
        background: white;
        padding: 20px;
        border-radius: 10px;
    }

    h3 {
        margin-top: 0;
    }

    .odds {
        display: flex;
        gap: 8px;
    }

    button {
        cursor: pointer;
        background: #e8eef7;
        border: 1px solid #cbd5e1;
        border-radius: 6px;
        padding: 12px;
        flex: 1;
        color: #192637;
    }

    button.chosen,
    aside button {
        background: #215bd7;
        color: white;
    }

    button:disabled {
        opacity: 0.5;
        cursor: default;
    }

    b {
        margin-left: 10px;
    }

    aside {
        margin-top: 20px;
    }

    input {
        padding: 12px;
        border: 1px solid #cbd5e1;
        border-radius: 6px;
        margin: 0 12px;
        width: 120px;
    }

    .online {
        color: #187747;
    }

    .offline {
        color: #b74121;
    }
</style>
