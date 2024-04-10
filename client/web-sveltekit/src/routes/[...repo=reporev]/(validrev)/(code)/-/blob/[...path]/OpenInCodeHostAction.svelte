<script lang="ts">
    import Tooltip from '$lib/Tooltip.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import type { OpenInCodeHostAction } from './OpenInCodeHostAction.gql'

    export let data: OpenInCodeHostAction
</script>

{#each data.externalURLs as { url, serviceKind } (url)}
    <Tooltip tooltip="Open in code host">
        <a href={url} target="_blank" rel="noopener noreferrer">
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
        color: var(--body-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>
