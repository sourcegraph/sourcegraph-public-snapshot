<script lang="ts">
    // @sg EnableRollout
    import { onMount } from 'svelte'

    import { beforeNavigate, goto } from '$app/navigation'
    import { resolveRoute } from '$app/paths'
    import { page } from '$app/stores'
    import Commit from '$lib/Commit.svelte'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import RepositoryRevPicker from '$lib/repo/RepositoryRevPicker.svelte'
    import Scroller, { type Capture } from '$lib/Scroller.svelte'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import { Alert } from '$lib/wildcard'
    import Button from '$lib/wildcard/Button.svelte'

    import type { PageData, Snapshot } from './$types'

    export let data: PageData

    export const snapshot: Snapshot<{
        scroller: Capture
        diffPagination: ReturnType<NonNullable<PageData['diffPagination']>['capture']>
        commitsPagination: ReturnType<NonNullable<PageData['commitsPagination']>['capture']>
        expandedDiffs: Array<[number, boolean]>
    }> = {
        capture() {
            return {
                scroller: scroller.capture(),
                diffPagination: data.diffPagination?.capture(),
                commitsPagination: data.commitsPagination?.capture(),
                expandedDiffs: expandedDiffsSnapshot,
            }
        },
        async restore(snapshot) {
            expandedDiffs = new Map(snapshot.expandedDiffs)
            await Promise.all([
                data.commitsPagination?.restore(snapshot.commitsPagination),
                data.diffPagination?.restore(snapshot.diffPagination),
            ])
            scroller.restore(snapshot.scroller)
        },
    }

    function handleSelect(baseRevision: string, headRevision: string): void {
        goto(
            resolveRoute('/[...repo=reporev]/-/compare/[...rangeSpec]', {
                repo: $page.params.repo,
                rangeSpec: `${baseRevision}...${headRevision}`,
            })
        )
    }

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('repo.compare', 'view')
    })

    beforeNavigate(event => {
        if (event.to?.route.id === $page.route.id && event.to.params?.['spec'] !== $page.params['spec']) {
            // Reset promise store when we navigate to a different commit range
            // Resolving the commit range can take a short but noticeable
            // amount of time and we don't want to show stale data while
            // the new data is being fetched.
            // Note: This is because resolving the commit range is
            // (intentionally) blocking page rendering.
            // todo: consider removing this when we have a page wide solution
            // for showing that a navigation is in progress.
            commitsPagination?.reset()
            diffPagination?.reset()
        }

        expandedDiffsSnapshot = Array.from(expandedDiffs.entries())
        expandedDiffs = new Map()
    })

    let scroller: Scroller
    let expandedDiffs = new Map<number, boolean>()
    let expandedDiffsSnapshot: Array<[number, boolean]> = []

    $: commitsPagination = data.commitsPagination
    $: diffPagination = data.diffPagination
    $: diffs = $diffPagination?.data
</script>

<svelte:head>
    <title>Compare - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<header>
    <h2>Compare changes across revisions</h2>
    <p>
        Select a revision or provide a <a
            href="https://git-scm.com/docs/git-rev-parse.html#_specifying_revisions"
            rel="noopener noreferrer"
            target="_blank">Git revspec</a
        > for more fine-grained comparisons.
    </p>
    <div class="controls">
        <span>
            <label for="page.compare.base-rev">base</label>
            <RepositoryRevPicker
                id="page.compare.base-rev"
                repoURL={data.repoURL}
                revision={data.base?.revision || ''}
                commitID={data.base?.commitID || ''}
                defaultBranch={data.defaultBranch}
                placement="bottom-start"
                onSelect={revision => handleSelect(revision, data.head?.revision || '')}
                getRepositoryBranches={data.getRepoBranches}
                getRepositoryCommits={data.getRepoCommits}
                getRepositoryTags={data.getRepoTags}
            />
        </span>
        <Icon icon={ILucideEllipsis} aria-hidden inline />
        <span>
            <label for="page.compare.head-rev">head</label>
            <RepositoryRevPicker
                id="page.compare.head-rev"
                repoURL={data.repoURL}
                revision={data.head?.revision || ''}
                commitID={data.head?.commitID || ''}
                defaultBranch={data.defaultBranch}
                placement="bottom-start"
                onSelect={revision => handleSelect(data.base?.revision || '', revision)}
                getRepositoryBranches={data.getRepoBranches}
                getRepositoryCommits={data.getRepoCommits}
                getRepositoryTags={data.getRepoTags}
            />
        </span>
    </div>
</header>

<hr />

<section>
    <Scroller bind:this={scroller} on:more={diffPagination?.fetchMore} margin={400}>
        <div class="wrapper">
            {#if data.error}
                <Alert variant="danger">{data.error.message}</Alert>
            {:else if !$commitsPagination || (!$commitsPagination.error && !$commitsPagination.loading && !$commitsPagination.data?.commits.length)}
                <Alert variant="info">No commits found between the selected revisions.</Alert>
            {:else if $commitsPagination}
                <div class="commits">
                    <h3>Commits</h3>
                    {#if $commitsPagination.data}
                        {@const hasNextPage = !!$commitsPagination.nextVariables}
                        {@const hasPrevPage = !!$commitsPagination.prevVariables}

                        <ul>
                            {#each $commitsPagination.data.commits as commit}
                                <li>
                                    <Commit {commit} />
                                </li>
                            {/each}
                        </ul>
                        <div class="controls" class:hidden={!(hasNextPage || hasPrevPage)}>
                            <Button variant="secondary">
                                <button
                                    slot="custom"
                                    let:buttonClass
                                    class={buttonClass}
                                    on:focus={hasPrevPage ? () => commitsPagination?.fetch('prev', true) : undefined}
                                    on:mouseover={hasPrevPage
                                        ? () => commitsPagination?.fetch('prev', true)
                                        : undefined}
                                    on:click={hasPrevPage ? () => commitsPagination?.fetch('prev') : undefined}
                                    disabled={!hasPrevPage || $commitsPagination.loading}
                                >
                                    <Icon icon={ILucideChevronLeft} inline aria-hidden /> Previous
                                </button>
                            </Button>
                            <Button variant="secondary">
                                <button
                                    slot="custom"
                                    let:buttonClass
                                    class={buttonClass}
                                    on:focus={hasNextPage ? () => commitsPagination?.fetch('next', true) : undefined}
                                    on:mouseover={hasNextPage
                                        ? () => commitsPagination?.fetch('next', true)
                                        : undefined}
                                    on:click={hasNextPage ? () => commitsPagination?.fetch('next') : undefined}
                                    disabled={!hasNextPage || $commitsPagination.loading}
                                >
                                    Next <Icon icon={ILucideChevronRight} inline aria-hidden />
                                </button>
                            </Button>
                        </div>
                    {:else if $commitsPagination.loading}
                        <LoadingSpinner />
                    {:else if $commitsPagination.error}
                        <Alert variant="warning"
                            >Unable to fetch commit information: {$commitsPagination.error.message}</Alert
                        >
                    {/if}
                </div>
            {/if}

            {#if !data.error}
                {#if diffs}
                    <ul class="diffs">
                        {#each diffs as node, index (index)}
                            <li>
                                <FileDiff
                                    fileDiff={node}
                                    expanded={expandedDiffs.get(index)}
                                    on:toggle={event => {
                                        expandedDiffs.set(index, event.detail.expanded)
                                        expandedDiffs = expandedDiffs
                                    }}
                                />
                            </li>
                        {/each}
                    </ul>
                {/if}
                {#if $diffPagination?.loading}
                    <LoadingSpinner />
                {:else if $diffPagination?.error}
                    <div class="error">
                        <Alert variant="danger">
                            Unable to fetch file diffs: {$diffPagination.error.message}
                        </Alert>
                    </div>
                {/if}
            {/if}
        </div>
    </Scroller>
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        overflow: hidden;
        height: 100%;
    }

    ul {
        list-style: none;
        margin: 0;
        padding: 0;
    }

    header .controls {
        display: flex;
        align-items: center;
        gap: 1rem;
    }

    .commits,
    .wrapper,
    header {
        padding: 0.5rem;
        margin: 0 auto;
        width: 100%;
        max-width: var(--viewport-xl);

        @media (--sm-breakpoint-up) {
            padding: 1rem;
        }
    }

    hr {
        width: 100%;
    }

    .commits {
        background-color: var(--color-bg-1);
        margin-bottom: 1rem;
        --avatar-size: 3rem;

        .controls {
            text-align: center;
        }

        li {
            padding: 0.5rem;

            + li {
                border-top: 1px solid var(--border-color);
            }
        }
    }

    .diffs {
        li {
            margin-bottom: 1rem;
        }
    }

    .hidden {
        display: none;
    }
</style>
