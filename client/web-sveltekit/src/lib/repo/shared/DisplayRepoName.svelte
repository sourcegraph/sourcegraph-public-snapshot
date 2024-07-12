<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import type { DisplayRepoName_ExternalLink } from './DisplayRepoName.gql'
    import { getIconForExternalService, inferExternalServiceKind } from './externalService'

    export let repoName: string
    export let externalLinks: DisplayRepoName_ExternalLink[] | undefined

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
    ><Tooltip tooltip={host ?? ''}><Icon icon={getIconForExternalService(kind)} inline /></Tooltip><DisplayPath
        path={displayName}>></DisplayPath
    ></span
>

<style lang="scss">
    span.root {
        display: contents;
        color: var(--text-body);
        font-family: var(--font-family-base);
        font-size: var(--font-size-base);
        font-weight: 500;

        :global([data-icon]) {
            margin-right: 0.375em;
            align-self: center;
            --icon-color: currentColor;
        }

        :global([data-path-container]) {
            color: var(--text-body);
            font-family: var(--font-family-base);
            font-size: var(--font-size-base);
            font-weight: 500;
            gap: 0;

            :global([data-slash]) {
                color: inherit;
            }
        }
    }
</style>
