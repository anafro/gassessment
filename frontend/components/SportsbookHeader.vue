<script setup lang="ts">
    import { ru } from '../locales/ru';
    import { money } from '../lib/format';
    defineProps<{ balance?: number }>();
</script>
<template>
    <header class="flex flex-wrap items-center justify-between gap-4">
        <h1 class="my-5 text-[28px] font-bold">{{ ru.appTitle }}</h1>
        <strong class="font-bold"
            >{{ ru.balance }}
            {{
                balance === undefined
                    ? ru.unavailable
                    : money(balance) + ' ' + ru.currency
            }}</strong
        >
    </header>
</template>
