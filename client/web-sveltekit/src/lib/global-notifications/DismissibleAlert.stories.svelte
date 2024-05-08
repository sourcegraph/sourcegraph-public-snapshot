<script lang="ts" context="module">
    import DismissibleAlert, { clearDismissedAlertsState_TEST_ONLY } from './DismissibleAlert.svelte'
    import { Story } from '@storybook/addon-svelte-csf'

    export const meta = {
        component: DismissibleAlert,
    }
</script>

<script lang="ts">
    let reset = 0
</script>

<Story name="Default">
    {#key reset}
        <DismissibleAlert partialStorageKey="a1" variant="info">This is an info alert</DismissibleAlert>
        <DismissibleAlert partialStorageKey="a2" variant="danger">This is a danger alert</DismissibleAlert>
        <DismissibleAlert partialStorageKey={null} variant="warning">This is a warning alert</DismissibleAlert>
    {/key}

    <hr />

    <button
        on:click={() => {
            clearDismissedAlertsState_TEST_ONLY('a1', 'a2')
            reset += 1
        }}>Reset dismissed alerts</button
    >
</Story>

<style lang="scss">
    hr {
        margin: 1rem 0;
    }
</style>
