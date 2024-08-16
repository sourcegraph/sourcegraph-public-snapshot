<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import RepositoryRevPicker from '$lib/repo/RepositoryRevPicker.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { Alert, Badge } from '$lib/wildcard'

    import type { PageData, Snapshot } from './$types'

    export let data: PageData

    // This tracks the number of commits that have been loaded and the current scroll
    // position, so both can be restored when the user refreshes the page or navigates
    // back to it.
    export const snapshot: Snapshot<{
        commits: ReturnType<typeof data.commitsQuery.capture>
        scroller: ScrollerCapture
    }> = {
        capture() {
            return {
                commits: commitsQuery.capture(),
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (get(navigating)?.type === 'popstate') {
                await commitsQuery?.restore(snapshot.commits)
            }
            scroller.restore(snapshot.scroller)
        },
    }

    function fetchMore() {
        commitsQuery?.fetchMore()
    }

    let scroller: Scroller

    $: commitsQuery = data.commitsQuery
    $: commits = $commitsQuery.data
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

<header>
    <h2>
        Commit History
        {#if data.path}
            in <code>{data.path}</code>
        {/if}
    </h2>
    <div>
        <RepositoryRevPicker
            repoURL={data.repoURL}
            revision={data.revision}
            commitID={data.resolvedRevision.commitID}
            defaultBranch={data.defaultBranch}
            placement="bottom-start"
            getRepositoryBranches={data.getRepoBranches}
            getRepositoryCommits={data.getRepoCommits}
            getRepositoryTags={data.getRepoTags}
        />
    </div>
</header>
<section>
    <Scroller bind:this={scroller} margin={600} on:more={fetchMore}>
        {#if commits}
            <ul class="commits">
                {#each commits as commit (commit.canonicalURL)}
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
        {#if $commitsQuery.fetching}
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

    h2 {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    header {
        border-bottom: 1px solid var(--border-color);
        div {
            width: min-content;
        }
    }

    header,
    ul.commits,
    .footer {
        max-width: var(--viewport-xl);
        width: 100%;
        margin: 0 auto;
        padding: 1rem;

        @media (--mobile) {
            padding: 0.5rem;
        }
    }

    ul {
        list-style: none;
    }

    ul.commits {
        --avatar-size: 2.5rem;
        padding-top: 0;

        > li {
            border-bottom: 1px solid var(--border-color);
            display: flex;
            padding: 0.5rem 0;
            gap: 1rem;

            @media (--mobile) {
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
