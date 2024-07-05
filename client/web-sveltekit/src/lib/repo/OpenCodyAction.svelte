<script lang="ts">
    import { CODY_SIDEBAR_ID } from '$lib/cody/CodySidebar.svelte'
    import Icon from '$lib/Icon.svelte'
    import { rightPanelOpen } from '$lib/repo/stores'
    import Tooltip from '$lib/Tooltip.svelte'

    function handleClick(): void {
        $rightPanelOpen = true
    }
</script>

<Tooltip tooltip="Open Cody chat">
    <button on:click={handleClick} aria-controls={CODY_SIDEBAR_ID} aria-expanded={$rightPanelOpen}>
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
