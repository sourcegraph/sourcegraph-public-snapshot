<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import Popover from '$lib/Popover.svelte'
    import { default as RepoPopover, fetchRepoPopoverData } from '$lib/repo/RepoPopover/RepoPopover.svelte'
    import DisplayRepoName from '$lib/repo/shared/DisplayRepoName.svelte'
    import { displayRepoName } from '$lib/shared'
    import { delay } from '$lib/utils'
    import Alert from '$lib/wildcard/Alert.svelte'

    export let repoName: string
    export let rev: string | undefined
    export let highlights: [number, number][] = []

    const client = getGraphQLClient()

    $: href = `/${repoName}${rev ? `@${rev}` : ''}`
    $: displayName = displayRepoName(repoName)
    $: if (displayName !== repoName) {
        // We only display part of the repository name, therefore we have to
        // adjust the match ranges for highlighting
        const delta = repoName.length - displayName.length
        highlights = highlights.map(([start, end]) => [start - delta, end - delta])
    }
</script>

<span class="root">
    <!-- #key is needed here to recreate the link because use:highlightRanges changes the DOM -->
    {#key highlights}
        <Popover showOnHover let:registerTrigger placement="bottom-start">
            <a class="repo-link" {href} use:highlightRanges={{ ranges: highlights }} use:registerTrigger>
                <DisplayRepoName {repoName} externalLinks={undefined} />
                {#if rev}
                    &nbsp;<small class="rev"> @ {rev}</small>
                {/if}
            </a>
            <svelte:fragment slot="content">
                {#await delay(fetchRepoPopoverData(client, repoName), 200) then data}
                    <RepoPopover {data} withHeader />
                {:catch error}
                    <Alert variant="danger" size="slim">{error}</Alert>
                {/await}
            </svelte:fragment>
        </Popover>
    {/key}
</span>

<style lang="scss">
    .root {
        --icon-color: currentColor;

        display: inline-flex;
        align-items: center;
        gap: 0.375rem;

        .repo-link {
            display: inline-flex;
            align-items: baseline;
            color: var(--text-body);
            font-weight: 500;

            .rev {
                color: var(--text-muted);
            }
        }
    }
</style>
