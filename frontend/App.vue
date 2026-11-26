<script setup lang="ts">
    import { computed } from 'vue';
    import { get } from '@vueuse/core';
    import SportsbookHeader from './components/SportsbookHeader.vue';
    import MatchLine from './components/MatchLine.vue';
    import BetSlip from './components/BetSlip.vue';
    import { useLiveSportsbook } from './composables/useLiveSportsbook';
    import { useBetSlip } from './composables/useBetSlip';

    const { matches, balance, connected, loadError } = useLiveSportsbook();
    const {
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
    } = useBetSlip({ matches, balance, connected });
    const statusMessage = computed(() => get(message) || get(loadError));
</script>
<template>
    <main class="mx-auto my-9 max-w-[950px] p-5">
        <SportsbookHeader :balance="balance" />
        <MatchLine
            :matches="matches"
            :connected="connected"
            :selected="selected"
            :disabled="selectionDisabled"
            @select="selectOutcome"
        />
        <BetSlip
            v-model:amount="amount"
            :selected="selected"
            :selected-match-name="selectedMatchName"
            :edit-disabled="editDisabled"
            :submit-disabled="submitDisabled"
            :busy="busy"
            :pending="!!pending"
            :message="statusMessage"
            @submit="submit"
        />
    </main>
</template>
