<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    // import { getGraphQLClient } from '$lib/graphql'
    import Popover from '$lib/Popover.svelte'
    // import { RepoPopover, type RepoPopoverResult } from '$lib/repo/RepoPopover/RepoPopover.gql'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { displayRepoName } from '$lib/shared'

    export let repoName: string
    export let rev: string | undefined
    export let highlights: [number, number][] = []
    // let repo: RepoPopoverResult | undefined

    /* const loadOnHover = async () => {
        const client = getGraphQLClient()
        const response = await client.query(RepoPopover, { repoName })
        repo = response.data
    }
 */
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
        <Popover let:registerTrigger let:toggle placement="bottom-start">
            <a
                class="repo-link"
                {href}
                use:highlightRanges={{ ranges: highlights }}
                use:registerTrigger
                on:mouseenter={() => {
                    toggle(true)
                    // loadOnHover()
                }}
                on:mouseleave={() => toggle(false)}
            >
                {displayRepoName(repoName)}
                {#if rev}
                    <small class="rev"> @ {rev}</small>
                {/if}
            </a>
            <div slot="content">{repoName}</div>
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
