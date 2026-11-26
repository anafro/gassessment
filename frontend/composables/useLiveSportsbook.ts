import { set } from '@vueuse/core';
import { onMounted, onUnmounted, ref } from 'vue';
import type { Match } from '../types';

/** Owns initial wallet loading and the live line connection for one mounted view. */
export function useLiveSportsbook() {
    const matches = ref<Match[]>([]);
    const balance = ref<number>();
    const connected = ref(false);
    const loadError = ref('');
    let events: EventSource | undefined;
    onUnmounted(() => events?.close());
    // Wallet loads once; EventSource reconnects automatically after transport errors.
    onMounted(async () => {
        try {
            const response = await fetch('/api/wallet');
            if (!response.ok) throw new Error();
            set(balance, (await response.json()).balance);
        } catch {
            set(loadError, 'Не удалось загрузить баланс');
        }
        events = new EventSource('/api/events');
        events.onmessage = (e) => {
            set(matches, JSON.parse(e.data));
            set(connected, true);
        };
        events.onerror = () => set(connected, false);
    });

    return { matches, balance, connected, loadError };
}
