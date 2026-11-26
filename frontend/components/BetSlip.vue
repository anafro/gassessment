<script setup lang="ts">
    import { ru } from '../locales/ru';
    import { money } from '../lib/format';
    import type { Selection } from '../types';
    defineProps<{
        selected?: Selection;
        selectedMatchName?: string;
        editDisabled: boolean;
        submitDisabled: boolean;
        busy: boolean;
        pending: boolean;
        message: string;
    }>();
    const amount = defineModel<string>('amount', { required: true });
    const emit = defineEmits<{ submit: [] }>();
</script>
<template>
    <aside class="mt-5 rounded-[10px] bg-white p-5">
        <h2 class="my-5 text-2xl font-bold">{{ ru.slip.title }}</h2>
        <p v-if="selected" class="my-4">
            {{ selectedMatchName }} · {{ ru.outcomes[selected.outcome] }} @
            {{ money(selected.odds) }}
        </p>
        <p v-else class="my-4">{{ ru.slip.selectOutcome }}</p>
        <label
            >{{ ru.slip.amount }}
            <input
                v-model="amount"
                inputmode="decimal"
                class="mx-3 w-30 rounded-md border border-slate-300 p-3"
                :disabled="editDisabled" /></label
        ><button
            type="button"
            class="cursor-pointer rounded-md border border-slate-300 bg-[#215bd7] p-3 text-white disabled:cursor-default disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            :disabled="submitDisabled"
            @click="emit('submit')"
        >
            {{
                busy
                    ? ru.slip.sending
                    : pending
                      ? ru.slip.retry
                      : ru.slip.submit
            }}
        </button>
        <p role="status" class="my-4">{{ message }}</p>
    </aside>
</template>
