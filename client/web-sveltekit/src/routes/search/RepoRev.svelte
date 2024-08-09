<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import Popover from '$lib/Popover.svelte'
    import DisplayRepoName from '$lib/repo/DisplayRepoName.svelte'
    import { default as RepoPopover, fetchRepoPopoverData } from '$lib/repo/RepoPopover/RepoPopover.svelte'
    import { delay } from '$lib/utils'
    import Alert from '$lib/wildcard/Alert.svelte'

    export let repoName: string
    export let rev: string | undefined
    export let highlights: [number, number][] = []

    const client = getGraphQLClient()

    $: href = `/${repoName}${rev ? `@${rev}` : ''}`
</script>

<!-- #key is needed here to recreate the link because use:highlightRanges changes the DOM -->
{#key highlights}
    <Popover showOnHover let:registerTrigger placement="bottom-start">
        <a class="repo-link" {href} use:highlightRanges={{ ranges: highlights }} use:registerTrigger>
            <DisplayRepoName {repoName} kind={undefined} />
            {#if rev}
                <small class="rev">&nbsp;@ {rev}</small>
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

<style lang="scss">
    .repo-link {
        display: flex;
        color: var(--body-color);
        font-weight: 500;
        align-items: baseline;
        :global([data-path-container]) {
            font-weight: 500;
        }
        .rev {
            color: var(--text-muted);
        }
    }
</style>
