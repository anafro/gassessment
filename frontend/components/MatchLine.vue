<script setup lang="ts">
    import { ru } from '../locales/ru';
    import { money } from '../lib/format';
    import type { Match, Selection } from '../types';
    defineProps<{
        matches: Match[];
        connected: boolean;
        disabled: boolean;
        selected?: Selection;
    }>();
    const emit = defineEmits<{ select: [selection: Selection] }>();
</script>
<template>
    <p class="my-4" :class="connected ? 'text-[#187747]' : 'text-[#b74121]'">
        {{ connected ? ru.line.connected : ru.line.disconnected }}
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
                    :disabled="disabled"
                    :class="
                        selected?.match === m.id && selected?.outcome === i
                            ? 'bg-[#215bd7] text-white'
                            : 'bg-[#e8eef7] text-[#192637]'
                    "
                    @click="
                        emit('select', { match: m.id, outcome: i, odds: odd })
                    "
                >
                    {{ ru.outcomes[i] }}
                    <b class="ml-2.5 font-bold">{{ money(odd) }}</b>
                </button>
            </div>
        </article>
    </section>
</template>
