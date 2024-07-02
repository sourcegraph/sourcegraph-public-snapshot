<script lang="ts">
    import { SourcegraphURL } from '@sourcegraph/common'

    import CodeExcerpt from '$lib/CodeExcerpt.svelte'
    import { observeIntersection } from '$lib/intersection-observer'
    import { pathHrefFactory } from '$lib/path'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { displayRepoName } from '$lib/shared'

    import type { ExplorePanel_Usage } from './ExplorePanel.gql'

    export let repo: string
    export let path: string
    export let usages: ExplorePanel_Usage[]
    export let scrollContainer: HTMLElement | null

    // TODO: remove all the usageRange! assertions once the backend is updated to
    // use a non-nullable type in the API. I've already confirmed that it should always
    // be non-null.
    //
    // FIXME: Assumes that all usages for a repo/path combo are at the same revision.
    $: revision = usages[0].usageRange!.revision

    let highlightedHTMLChunks: string[][] | undefined
    let visible = false
    $: if (visible) {
        fetchFileRangeMatches({
            result: {
                repository: repo,
                commit: revision,
                path: path,
            },
            ranges: usages.map(usage => ({
                startLine: usage.usageRange!.range.start.line,
                endLine: usage.usageRange!.range.end.line + 1,
            })),
        })
            .then(result => {
                highlightedHTMLChunks = result
            })
            .catch(err => console.error('Failed to fetch highlighted usages', err))
    }

    function hrefForUsage(usage: ExplorePanel_Usage): string {
        const { repository, revision, path, range } = usage.usageRange!
        // TODO: omit revision if navigating to the default branch.
        return SourcegraphURL.from(`${repository}@${revision}/-/blob/${path}`)
            .setLineRange({
                line: range.start.line + 1,
                character: range.start.character + 1,
                endLine: range.end.line + 1,
                endCharacter: range.end.character + 1,
            })
            .toString()
    }

    $: usageExcerpts = usages.map((usage, index) => ({
        startLine: usage.usageRange!.range.start.line,
        matches: [
            {
                startLine: usage.usageRange!.range.start.line,
                startCharacter: usage.usageRange!.range.start.character,
                endLine: usage.usageRange!.range.end.line,
                endCharacter: usage.usageRange!.range.end.character,
            },
        ],
        plaintextLines: [usage.surroundingContent],
        highlightedHTMLRows: highlightedHTMLChunks?.[index],
        href: hrefForUsage(usage),
    }))
</script>

<div
    class="root"
    use:observeIntersection={scrollContainer}
    on:intersecting={event => (visible = visible || event.detail)}
>
    <div class="header">
        <CodeHostIcon repository={repo} />
        <span class="repo-name"><DisplayPath path={displayRepoName(repo)} /></span>
        <span class="interpunct">⋅</span>
        <DisplayPath
            {path}
            pathHref={pathHrefFactory({
                repoName: repo,
                revision: revision,
                fullPath: path,
                fullPathType: 'blob',
            })}
            showCopyButton
        />
    </div>
    <div>
        {#each usageExcerpts as excerpt}
            <a href={excerpt.href}>
                <CodeExcerpt
                    collapseWhitespace
                    startLine={excerpt.startLine}
                    plaintextLines={excerpt.plaintextLines}
                    matches={excerpt.matches}
                    highlightedHTMLRows={excerpt.highlightedHTMLRows}
                />
            </a>
        {/each}
    </div>
</div>

<style lang="scss">
    .root {
        :global([data-copy-button]) {
            opacity: 0;
            transition: opacity 0.2s;
        }
        &:is(:hover, :focus-within) :global([data-copy-button]) {
            opacity: 1;
        }
    }

    .header {
        display: flex;
        gap: 0.5rem;
        background-color: var(--secondary-4);
        padding: 0.375rem 0.75rem;
        border-bottom: 1px solid var(--border-color);
        align-items: center;

        position: sticky;
        top: 0;

        --icon-color: currentColor;
        .repo-name {
            :global([data-path-container]) {
                font-family: var(--font-family-base);
                font-weight: 500;
                font-size: var(--font-size-small);
                gap: 0.25rem;
            }

            :global([data-path-item]) {
                color: var(--text-title);
            }
        }

        .interpunct {
            color: var(--text-disabled);
        }
    }

    a {
        display: block;
        padding: 0.25rem 0.5rem;
        border-bottom: 1px solid var(--border-color);
        text-decoration: none;
        &:hover,
        &:focus {
            background-color: var(--secondary-4);
        }
    }
</style>
