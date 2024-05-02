<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import Popover from '$lib/Popover.svelte'
    import RepoPopover from '$lib/repo/RepoPopover/RepoPopover.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { displayRepoName } from '$lib/shared'

    import { RepoPopoverQuery } from '../layout.gql'

    export let repoName: string
    export let rev: string | undefined
    export let highlights: [number, number][] = []

    const fetchPopoverInfo = async (repo: string) => {
        const client = getGraphQLClient()
        const popoverInfo = await client.query(RepoPopoverQuery, { repoName })
        if (popoverInfo.data) {
            return popoverInfo.data.repository
        }
        console.error('Failed to fetch popover info for', repo)
        throw new Error('Failed to fetch popover info')
    }

    $: popoverInfo = fetchPopoverInfo(repoName)
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
    <CodeHostIcon repository={repoName} />
    <!-- #key is needed here to recreate the link because use:highlightRanges changes the DOM -->
    {#key highlights}
        <Popover showOnHover let:registerTrigger placement="bottom-start" useDefaultBorder={false}>
            <a class="repo-link" {href} use:highlightRanges={{ ranges: highlights }} use:registerTrigger>
                {displayRepoName(repoName)}
                {#if rev}
                    <small class="rev"> @ {rev}</small>
                {/if}
            </a>
            <div slot="content">
                {#await popoverInfo then popoverInfo}
                    {#if popoverInfo !== null}
                        <RepoPopover repo={popoverInfo} withHeader={true} />
                    {/if}
                {/await}
            </div>
        </Popover>
    {/key}
</span>

<style lang="scss">
    .root {
        display: inline-flex;
        align-items: center;
        gap: 0.375rem;

        .repo-link {
            align-self: flex-start;
            color: var(--text-body);
            .rev {
                color: var(--text-muted);
            }
        }
    }
</style>
