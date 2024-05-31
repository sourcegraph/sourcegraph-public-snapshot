<script lang="ts" context="module">
    import { writable } from 'svelte/store'

    export const lineWrap = writable<boolean>(false)
</script>

<script lang="ts">
    import Icon2 from '$lib/Icon2.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    $: tooltip = ($lineWrap ? 'Disable' : 'Enable') + ' wrapping long lines'
</script>

<Tooltip {tooltip}>
    <button on:click={() => lineWrap.update(wrap => !wrap)}
        ><Icon2 icon={$lineWrap ? ILucideWrapText : ILucideText} inline /></button
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
