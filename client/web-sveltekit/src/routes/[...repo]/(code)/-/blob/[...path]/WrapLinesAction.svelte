<script lang="ts" context="module">
    import { writable } from 'svelte/store'

    export const lineWrap = writable<boolean>(false)
</script>

<script lang="ts">
    import { mdiWrap, mdiWrapDisabled } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    $: tooltip = ($lineWrap ? 'Disable' : 'Enable') + ' wrapping long lines'
</script>

<Tooltip {tooltip}>
    <button on:click={() => lineWrap.update(wrap => !wrap)}
        ><Icon svgPath={$lineWrap ? mdiWrap : mdiWrapDisabled} inline /></button
    >
</Tooltip>

<style lang="scss">
    button {
        margin: 0;
        padding: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
    }
</style>
