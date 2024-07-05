<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { rightPanelOpen } from '$lib/repo/stores'
    import Tooltip from '$lib/Tooltip.svelte'

    function handleClick(): void {
        $rightPanelOpen = true
    }
</script>

<Tooltip tooltip="Open Cody chat">
    <button on:click={handleClick}>
        <Icon icon={ISgCody} />
        <span data-action-label>Cody</span>
    </button>
</Tooltip>

<style lang="scss">
    button {
        all: unset;
        white-space: nowrap;
        cursor: pointer;
    }
</style>
