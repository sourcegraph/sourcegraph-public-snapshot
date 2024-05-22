<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import Commit from '$lib/Commit.svelte'
    import { pluralize } from '$lib/common'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import Alert from '$lib/wildcard/Alert.svelte'
    import Badge from '$lib/wildcard/Badge.svelte'
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    import type { PageData, Snapshot } from './$types'

    interface Capture {
        scroll: ScrollerCapture
        diffCount: number
        expandedDiffs: Array<[number, boolean]>
    }

    export let data: PageData

    export const snapshot: Snapshot<Capture> = {
        capture: () => ({
            scroll: scroller.capture(),
            diffCount: diffs?.nodes.length ?? 0,
            expandedDiffs: Array.from(expandedDiffs.entries()),
        }),
        restore: async capture => {
            expandedDiffs = new Map(capture.expandedDiffs)
            if (capture?.diffCount !== undefined && get(navigating)?.type === 'popstate') {
                await data.diff?.restore(result => {
                    const count = result.data?.repository?.comparison.fileDiffs.nodes.length
                    return !!count && count < capture.diffCount
                })
            }
            scroller.restore(capture.scroll)
        },
    }

    let scroller: Scroller
    let expandedDiffs = new Map<number, boolean>()

    $: diffQuery = data.diff
    $: diffs = $diffQuery?.data?.repository?.comparison.fileDiffs ?? null
</script>

<svelte:head>
    <title>Commit: {data.commit?.subject ?? ''} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if data.commit}
        <Scroller bind:this={scroller} margin={600} on:more={data.diff?.fetchMore}>
            <div class="header">
                <div class="info"><Commit commit={data.commit} alwaysExpanded /></div>
                <div class="parents">
                    <span>Commit:</span>
                    <Badge variant="secondary"><code>{data.commit.abbreviatedOID}</code></Badge>&nbsp;<CopyButton
                        value={data.commit.abbreviatedOID}
                    />
                    <br />
                    <span>{pluralize('Parent', data.commit.parents.length)}:</span>
                    {#each data.commit.parents as parent}
                        <Badge variant="link"><a href={parent.canonicalURL}>{parent.abbreviatedOID}</a></Badge
                        >&nbsp;<CopyButton value={parent.abbreviatedOID} />{' '}
                    {/each}
                    <br />
                    <ul>
                        <li>
                            <a href="/{data.repoName}@{data.commit.oid}"
                                >Browse files at <Badge variant="link">{data.commit.abbreviatedOID}</Badge></a
                            >
                        </li>
                        {#each data.commit.externalURLs as { url, serviceKind }}
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
                </div>
            </div>
            <hr />
            {#if !$diffQuery?.restoring && diffs}
                <ul class="diffs">
                    {#each diffs.nodes as node, index}
                        <li>
                            <FileDiff
                                fileDiff={node}
                                expanded={expandedDiffs.get(index)}
                                on:toggle={event => expandedDiffs.set(index, event.detail.expanded)}
                            />
                        </li>
                    {/each}
                </ul>
            {/if}
            {#if $diffQuery?.fetching || $diffQuery?.restoring}
                <LoadingSpinner />
            {:else if $diffQuery?.error}
                <div class="error">
                    <Alert variant="danger">
                        Unable to fetch file diffs: {$diffQuery.error.message}
                    </Alert>
                </div>
            {/if}
        </Scroller>
    {/if}
</section>

<style lang="scss">
    section {
        overflow: auto;
    }

    .header {
        display: flex;
        margin: 1rem;

        ul {
            all: unset;
            list-style: none;
        }
    }

    .parents {
        span,
        a {
            vertical-align: middle;
        }
    }

    .info {
        flex: 1;
        --avatar-size: 2.5rem;
    }

    code {
        font-family: monospace;
        font-size: inherit;
    }

    .error,
    ul.diffs {
        padding: 1rem;
    }

    ul.diffs {
        // Removes globally set margin
        margin: 0;
        list-style: none;

        li:not(:last-child) {
            margin-bottom: 1rem;
        }
    }
</style>
