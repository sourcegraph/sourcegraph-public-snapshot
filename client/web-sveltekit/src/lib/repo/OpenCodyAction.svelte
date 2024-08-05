<script lang="ts">
    import { CODY_SIDEBAR_ID } from '$lib/cody/CodySidebar.svelte'
    import Icon from '$lib/Icon.svelte'
    import { rightSidePanelOpen } from '$lib/repo/stores'
    import Tooltip from '$lib/Tooltip.svelte'

    function handleClick(): void {
        $rightSidePanelOpen = true
    }
</script>

{#if !$rightSidePanelOpen}
    <Tooltip tooltip="Open Cody chat">
        <button on:click={handleClick} aria-controls={CODY_SIDEBAR_ID} aria-expanded={$rightSidePanelOpen}>
            <Icon icon={ISgCody} />
            <span data-action-label>Cody</span>
        </button>
    </Tooltip>
{/if}

<style lang="scss">
    button {
        all: unset;
        white-space: nowrap;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 0.375rem;
        color: var(--text-body);
    }
</style>
