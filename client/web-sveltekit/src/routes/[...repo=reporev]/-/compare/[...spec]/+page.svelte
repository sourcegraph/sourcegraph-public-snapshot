<script lang="ts">
    // @sg EnableRollout
    import { onMount } from 'svelte'

    import { goto } from '$app/navigation'
    import { resolveRoute } from '$app/paths'
    import { page } from '$app/stores'
    import Commit from '$lib/Commit.svelte'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import Scroller, { type Capture } from '$lib/Scroller.svelte'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import { createPromiseStore } from '$lib/utils'
    import { Alert } from '$lib/wildcard'
    import Button from '$lib/wildcard/Button.svelte'

    import RepositoryRevPicker from '$lib/repo/RepositoryRevPicker.svelte'

    import type { PageData, Snapshot } from './$types'

    export let data: PageData

    export const snapshot: Snapshot<{
        scroller: Capture
        diffs: ReturnType<NonNullable<typeof data.diffQuery>['capture']>
    }> = {
        capture() {
            return {
                scroller: scroller.capture(),
                diffs: data.diffQuery?.capture(),
            }
        },
        async restore({ scroller: scrollerData, diffs: diffsData }) {
            await data.diffQuery?.restore(diffsData)
            scroller.restore(scrollerData)
        },
    }

    function handleSelect(baseRevision: string, headRevision: string): void {
        goto(
            resolveRoute('/[...repo=reporev]/-/compare/[...spec]', {
                repo: $page.params.repo,
                spec: `${baseRevision}...${headRevision}`,
            })
        )
    }

    function handleCommitPage(page: number): void {
        const url = new URL($page.url)
        url.searchParams.set('p', page.toString())
        goto(url)
    }

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('repo.compare', 'view')
    })

    let scroller: Scroller
    let expandedDiffs = new Map<number, boolean>()
    const commits = createPromiseStore<Awaited<typeof data.commits>>()

    $: commits.set(data.commits)
    $: diffQuery = data.diffQuery
    $: diffs = $diffQuery?.data
</script>

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
            <label for="base-rev">base</label>
            <RepositoryRevPicker
                id="base-rev"
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
            <label for="head-rev">head</label>
            <RepositoryRevPicker
                id="head-rev"
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
    <Scroller bind:this={scroller} on:more={diffQuery?.fetchMore} margin={400}>
        <div class="wrapper">
            {#if data.error}
                <Alert variant="danger">{data.error.message}</Alert>
            {:else if !$commits.pending && !$commits.error && (!$commits.value || $commits.value.commits.length === 0)}
                <Alert variant="info">No commits found between the selected revisions.</Alert>
            {:else}
                <div class="commits">
                    <h3>Commits</h3>
                    {#if $commits.value}
                        {@const previousPage = $commits.value.previousPage}
                        {@const nextPage = $commits.value.nextPage}

                        <ul>
                            {#each $commits.value.commits as commit}
                                <li>
                                    <Commit {commit} />
                                </li>
                            {/each}
                        </ul>
                        {#if previousPage || nextPage}
                            <div class="controls">
                                <Button
                                    variant="secondary"
                                    on:click={previousPage ? () => handleCommitPage(previousPage) : undefined}
                                    disabled={!previousPage || $commits.pending}
                                >
                                    <Icon icon={ILucideChevronLeft} inline aria-hidden /> Previous
                                </Button>
                                <Button
                                    variant="secondary"
                                    on:click={nextPage ? () => handleCommitPage(nextPage) : undefined}
                                    disabled={!nextPage || $commits.pending}
                                >
                                    Next <Icon icon={ILucideChevronRight} inline aria-hidden />
                                </Button>
                            </div>
                        {/if}
                    {:else if $commits.pending}
                        <LoadingSpinner />
                    {:else if $commits.error}
                        <Alert variant="warning">Unable to fetch commit information: {$commits.error.message}</Alert>
                    {/if}
                </div>
            {/if}

            {#if !data.error}
                <ul class="diffs">
                    {#each diffs ?? [] as node, index}
                        <li>
                            <FileDiff
                                fileDiff={node}
                                expanded={expandedDiffs.get(index)}
                                on:toggle={event => expandedDiffs.set(index, event.detail.expanded)}
                            />
                        </li>
                    {/each}
                </ul>
                {#if $diffQuery?.fetching}
                    <LoadingSpinner />
                {:else if $diffQuery?.error}
                    <div class="error">
                        <Alert variant="danger">
                            Unable to fetch file diffs: {$diffQuery.error.message}
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
</style>
