<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import { pluralize } from '$lib/common'
    import { GitRefType } from '$lib/graphql-types'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReferencesTable from '$lib/repo/GitReferencesTable.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { Alert, Button, Input } from '$lib/wildcard'

    import type { PageData, Snapshot } from './$types'

    export let data: PageData

    export const snapshot: Snapshot<{ tags: ReturnType<typeof data.tagsQuery.capture>; scroller: ScrollerCapture }> = {
        capture() {
            return {
                tags: data.tagsQuery.capture(),
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (snapshot?.tags && get(navigating)?.type === 'popstate') {
                await data.tagsQuery?.restore(snapshot.tags)
            }
            scroller.restore(snapshot.scroller)
        },
    }

    let scroller: Scroller

    $: query = data.query
    $: tagsQuery = data.tagsQuery
    $: tags = $tagsQuery.data
</script>

<svelte:head>
    <title>Tags - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <form method="GET">
        <Input type="search" name="query" placeholder="Search tags" value={query} autofocus />
        <Button variant="primary" type="submit">Search</Button>
    </form>
    <Scroller bind:this={scroller} margin={600} on:more={tagsQuery.fetchMore}>
        <div class="main">
            {#if tags && tags.nodes.length > 0}
                <GitReferencesTable references={tags.nodes} referenceType={GitRefType.GIT_TAG} />
            {/if}
            <div>
                {#if $tagsQuery.fetching}
                    <LoadingSpinner />
                {:else if $tagsQuery.error}
                    <Alert variant="danger">
                        Unable to load tags: {$tagsQuery.error.message}
                    </Alert>
                {:else if !tags || tags.nodes.length === 0}
                    <Alert variant="info">No tags found</Alert>
                {/if}
            </div>
        </div>
    </Scroller>
    {#if tags && tags.nodes.length > 0}
        <div class="footer">
            {tags.totalCount}
            {pluralize('tag', tags.totalCount)} total
            {#if tags.totalCount > tags.nodes.length}
                (showing {tags.nodes.length})
            {/if}
        </div>
    {/if}
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
        gap: 1rem;
        padding: 0.5rem 0;

        :global([data-scroller]) {
            display: flex;
            flex-direction: column;
        }
    }

    form {
        display: flex;
        gap: 1rem;

        :global([data-input-container]) {
            flex: 1;
        }
    }

    form,
    .footer,
    .main {
        align-self: center;
        max-width: var(--viewport-lg);
        width: 100%;
        padding: 0 1rem;
    }

    @media (--mobile) {
        .main {
            padding: 0;
        }
    }

    .footer {
        color: var(--text-muted);
        // Unset `div` width: 100% to allow the footer to be centered
        width: initial;
    }
</style>
