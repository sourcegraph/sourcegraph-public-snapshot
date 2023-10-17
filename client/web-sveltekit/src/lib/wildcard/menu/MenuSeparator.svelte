<script lang="ts">
    import { getContext } from './Menu.svelte'

    const { separator } = getContext()
</script>

<div {...$separator} use:separator />

<style lang="scss">
    div {
        height: 0;
        margin: 0.25rem 0;
        overflow: hidden;
        border-top: 1px solid var(--dropdown--separator-color, var(--border-color));
    }
</style>
