<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'
    import type { Commits } from './page.gql'
    import Paginator from '$lib/Paginator.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'

    export let data: PageData

    const { pending, latestValue: commits, set } = createPromiseStore<Promise<Commits>>()
    $: set(data.deferred.commits)

    // This is a hack to make backward pagination work.
    $: hasPreviousPage = data.afterCursor && +data.afterCursor > 0
    $: previousEndCursor = data.afterCursor ? String(+data.afterCursor - data.pageSize) : null
</script>

<svelte:head>
    <title>Commits - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if $pending && !$commits}
        <div class="loader">
            <LoadingSpinner />
        </div>
    {:else if $commits}
        <div class="commits">
            <ul>
                {#each $commits.nodes as commit (commit.canonicalURL)}
                    <li><Commit {commit} /></li>
                {/each}
            </ul>
        </div>
        <div class="paginator">
            <Paginator
                disabled={$pending}
                pageInfo={{
                    ...$commits.pageInfo,
                    hasPreviousPage,
                    previousEndCursor,
                }}
                showLastpageButton={false}
            />
            <div class="loader" class:visible={$pending}>
                <LoadingSpinner />
            </div>
        </div>
    {/if}
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        flex: 1;
        min-height: 0;

        > .loader {
            flex: 1;
            display: flex;
        }
    }

    .commits {
        overflow-y: auto;
        flex: 1;
    }

    ul {
        list-style: none;
        padding: 1rem;
        max-width: var(--viewport-xl);
        margin: 0 auto;
        --avatar-size: 2.5rem;
    }

    .paginator {
        flex: 0 0 auto;
        margin: 1rem auto;
        display: flex;
        align-items: center;

        .loader {
            margin-left: 1rem;
            visibility: hidden;

            &.visible {
                visibility: visible;
            }
        }
    }

    li {
        border-bottom: 1px solid var(--border-color);
        padding: 0.5rem 0;

        &:last-child {
            border: none;
        }
    }
</style>
