<script lang="ts">
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'
    import Tooltip from '$lib/Tooltip.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import type { OpenInCodeHostAction } from './OpenInCodeHostAction.gql'

    export let data: OpenInCodeHostAction

    function handleOpenCodeHostClick(): void {
        SVELTE_LOGGER.log(SVELTE_TELEMETRY_EVENTS.GoToCodeHost)
    }
</script>

{#each data.externalURLs as { url, serviceKind } (url)}
    <Tooltip tooltip="Open in code host">
        <a href={url} target="_blank" rel="noopener noreferrer" on:click={handleOpenCodeHostClick}>
            {#if serviceKind}
                <CodeHostIcon repository={serviceKind} disableTooltip />
                <span data-action-label>
                    {getHumanNameForCodeHost(serviceKind)}
                </span>
            {:else}
                Code host
            {/if}
        </a>
    </Tooltip>
{/each}

<style lang="scss">
    a {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        color: var(--text-body);
        text-decoration: none;
        white-space: nowrap;

        &:hover {
            color: var(--text-title);
        }
    }
</style>
