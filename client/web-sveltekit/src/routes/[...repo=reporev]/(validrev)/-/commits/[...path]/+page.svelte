<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { Alert, Badge } from '$lib/wildcard'

    import type { PageData, Snapshot } from './$types'
    import type { CommitsPage_GitCommitConnection } from './page.gql'

    export let data: PageData

    // This tracks the number of commits that have been loaded and the current scroll
    // position, so both can be restored when the user refreshes the page or navigates
    // back to it.
    export const snapshot: Snapshot<{ commitCount: number; scroller: ScrollerCapture }> = {
        capture() {
            return {
                commitCount: commits?.nodes.length ?? 0,
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (snapshot?.commitCount !== undefined && get(navigating)?.type === 'popstate') {
                await commitsQuery?.restore(result => {
                    const count = result.data?.repository?.commit?.ancestors.nodes?.length
                    return !!count && count < snapshot.commitCount
                })
            }
            scroller.restore(snapshot.scroller)
        },
    }

    function fetchMore() {
        commitsQuery?.fetchMore()
    }

    let scroller: Scroller
    let commits: CommitsPage_GitCommitConnection | null = null

    $: commitsQuery = data.commitsQuery
    // We conditionally check for the ancestors field to be able to show
    // previously loaded commits when an error occurs while fetching more commits.
    $: if ($commitsQuery?.data?.repository?.commit?.ancestors) {
        commits = $commitsQuery.data.repository.commit.ancestors
    }
    $: pageTitle = (() => {
        const parts = ['Commits']
        if (data.path) {
            parts.push(data.path)
        }
        parts.push(data.displayRepoName, 'Sourcegraph')
        return parts.join(' - ')
    })()
</script>

<svelte:head>
    <title>{pageTitle}</title>
</svelte:head>

{#if data.path}
    <h2>Commits in <code>{data.path}</code></h2>
{/if}
<section>
    <Scroller bind:this={scroller} margin={600} on:more={fetchMore}>
        {#if !$commitsQuery.restoring && commits}
            <ul class="commits">
                {#each commits.nodes as commit (commit.canonicalURL)}
                    <li>
                        <div class="commit">
                            <Commit {commit} />
                        </div>
                        <ul class="actions">
                            <li>
                                <Badge variant="link">
                                    <a href={commit.canonicalURL} title="View commit">{commit.abbreviatedOID}</a>
                                </Badge>
                            </li>
                            <li><a href="/{data.repoName}@{commit.oid}">Browse files</a></li>
                            {#each commit.externalURLs as { url, serviceKind }}
                                <li>
                                    <a href={url}>
                                        View on
                                        {#if serviceKind}
                                            <CodeHostIcon repository={serviceKind} disableTooltip />
                                            {getHumanNameForCodeHost(serviceKind)}
                                        {:else}
                                            code host
                                        {/if}
                                    </a>
                                </li>
                            {/each}
                        </ul>
                    </li>
                {:else}
                    <li>
                        <Alert variant="info">No commits found</Alert>
                    </li>
                {/each}
            </ul>
        {/if}
        {#if $commitsQuery.fetching || $commitsQuery.restoring}
            <div class="footer">
                <LoadingSpinner />
            </div>
        {:else if !$commitsQuery.fetching && $commitsQuery.error}
            <div class="footer">
                <Alert variant="danger">
                    Unable to fetch commits: {$commitsQuery.error.message}
                </Alert>
            </div>
        {/if}
    </Scroller>
</section>

<style lang="scss">
    section {
        overflow: hidden;
    }

    ul {
        margin: 0;
        padding: 0;
    }

    h2,
    ul.commits,
    .footer {
        max-width: var(--viewport-xl);
        width: 100%;
        margin: 0 auto;
        padding: 0.5rem;

        @media (--sm-breakpoint-up) {
            padding: 1rem;
        }
    }

    ul {
        list-style: none;
    }

    ul.commits {
        --avatar-size: 2.5rem;

        > li {
            border-bottom: 1px solid var(--border-color);
            display: flex;
            padding: 0.5rem 0;
            gap: 1rem;

            @media (--xs-breakpoint-down) {
                display: block;

                .actions {
                    display: flex;
                    gap: 0.5rem;
                    margin-top: 0.5rem;

                    li:not(:last-child)::after {
                        content: '•';
                        padding-left: 0.5rem;
                        color: var(--text-muted);
                    }
                }
            }

            .commit {
                flex: 1;
                min-width: 0;
            }

            .actions {
                --icon-color: currentColor;

                flex-shrink: 0;
            }

            &:last-child {
                border: none;
            }
        }
    }

    .footer {
        &:not(:first-child) {
            border-top: 1px solid var(--border-color);
        }
    }
</style>
