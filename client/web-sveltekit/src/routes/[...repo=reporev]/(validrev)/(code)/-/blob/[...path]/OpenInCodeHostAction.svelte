<script lang="ts">
    import type { LineOrPositionOrRange } from '$lib/common'
    import CodeHostIcon from '$lib/repo/codehost/CodeHostIcon.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import { getExternalURL } from '$lib/repo/url'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import Tooltip from '$lib/Tooltip.svelte'

    import type { OpenInCodeHostAction } from './OpenInCodeHostAction.gql'

    export let data: OpenInCodeHostAction
    export let lineOrPosition: LineOrPositionOrRange | undefined = undefined

    function handleOpenCodeHostClick(): void {
        TELEMETRY_RECORDER.recordEvent('repo.goToCodeHost', 'click')
    }
</script>

{#each data.externalURLs as externalLink (externalLink.url)}
    <Tooltip tooltip="Open in code host">
        <a
            href={getExternalURL({ externalLink, lineOrPosition })}
            target="_blank"
            rel="noopener noreferrer"
            on:click={handleOpenCodeHostClick}
        >
            {#if externalLink.serviceKind}
                <CodeHostIcon kind={externalLink.serviceKind} inline />
                <span data-action-label>
                    {getHumanNameForCodeHost(externalLink.serviceKind)}
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
