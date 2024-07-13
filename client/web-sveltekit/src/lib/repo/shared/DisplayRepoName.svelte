<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import type { DisplayRepoName_ExternalLink } from './DisplayRepoName.gql'
    import { getIconForExternalService, inferExternalServiceKind } from './externalService'

    export let repoName: string
    export let externalLinks: DisplayRepoName_ExternalLink[] | undefined
    export let disableTooltip: boolean = false

    function linkHost(link: DisplayRepoName_ExternalLink): string {
        return new URL(link.url).hostname
    }

    // Find an external link that matches the repo name (if it exists).
    // Fall back to arbitrarily selecting the first external link (if it exists).
    $: link = externalLinks?.find(link => repoName.startsWith(`${linkHost(link)}/`)) ?? externalLinks?.at(0)
    $: host = link
        ? `${linkHost(link)}`
        : repoName
              .match(/^[^\/]*\.[^\/]*\//)
              ?.at(0)
              ?.slice(0, -1)
    $: displayName = host ? repoName.slice(host.length + 1) : repoName
    $: kind = link?.serviceKind ?? inferExternalServiceKind(repoName)
</script>

<span class="root"
    ><Tooltip tooltip={disableTooltip ? '' : host ?? ''}><Icon icon={getIconForExternalService(kind)} inline /></Tooltip
    ><DisplayPath path={displayName}>></DisplayPath></span
>

<style lang="scss">
    span.root {
        display: flex;

        :global([data-icon]) {
            flex: none;
            margin-right: 0.375em;
            align-self: center;
            --icon-color: currentColor;
        }

        :global([data-path-container]) {
            color: inherit;
            font-size: inherit;
            font-weight: inherit;
            gap: 0;
            font-family: var(--font-family-base);

            :global([data-path-item]) {
                color: inherit;
            }

            :global([data-slash]) {
                color: inherit;
            }
        }
    }
</style>
