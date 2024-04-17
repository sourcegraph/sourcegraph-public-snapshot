<script lang="ts">
    import { onMount } from 'svelte'

    import { highlightRanges } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Popover from '$lib/Popover.svelte'
    import { RepoPopover as RepoPopoverQuery, type RepoPopoverResult } from '$lib/repo/RepoPopover/RepoPopover.gql.ts'
    import RepoPopover from '$lib/repo/RepoPopover/RepoPopover.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { displayRepoName } from '$lib/shared'

    export let repoName: string
    export let rev: string | undefined
    export let highlights: [number, number][] = []

    let repo: RepoPopoverResult

    onMount(async () => {
        const client = getGraphQLClient()
        const response = await client.query(RepoPopoverQuery, { repoName })
        if (response.data) {
            repo = response.data
        }
    })

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
        <Popover showOnHover let:registerTrigger placement="bottom-start">
            <a class="repo-link" {href} use:highlightRanges={{ ranges: highlights }} use:registerTrigger>
                {displayRepoName(repoName)}
                {#if rev}
                    <small class="rev"> @ {rev}</small>
                {/if}
            </a>
            <div slot="content">
                {#if repo}
                    <RepoPopover repo={repo.repository} withHeader={true} />
                {:else}
                    <div class="loading">
                        <LoadingSpinner />
                    </div>
                {/if}
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

    .loading {
        display: flex;
        flex-flow: column nowrap;
        align-items: center;
        justify-content: flex-start;
        width: 480px;
        height: 200px;
        gap: 0.5rem;
        padding: 1.5rem 1rem;
    }
</style>
