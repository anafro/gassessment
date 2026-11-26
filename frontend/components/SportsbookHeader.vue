<script setup lang="ts">
    import { money } from '../lib/format';
    defineProps<{ balance?: number }>();
</script>
<template>
    <header class="flex flex-wrap items-center justify-between gap-4">
        <h1 class="my-5 text-[28px] font-bold">Мини-спортсбук</h1>
        <strong class="font-bold"
            >Баланс:
            {{ balance === undefined ? '—' : money(balance) + ' ₽' }}</strong
        >
    </header>
</template>
