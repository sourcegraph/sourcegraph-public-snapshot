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

    export let repository: string
    export let path: string
    export let usages: ExplorePanel_Usage[]
    export let scrollContainer: HTMLElement | undefined

    // FIXME: Assumes that all usages for a repo/path combo are at the same revision.
    $: revision = usages[0].usageRange.revision

    let highlightedHTMLChunks: string[][] | undefined
    let visible = false
    $: if (visible) {
        fetchFileRangeMatches({
            result: {
                repository,
                commit: revision,
                path: path,
            },
            ranges: usages.map(usage => ({
                startLine: usage.usageRange.range.start.line,
                endLine: usage.usageRange.range.end.line + 1,
            })),
        })
            .then(result => {
                highlightedHTMLChunks = result
            })
            .catch(err => console.error('Failed to fetch highlighted usages', err))
    }

    function hrefForUsage(usage: ExplorePanel_Usage): string {
        const { repository, revision, path, range } = usage.usageRange
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
        startLine: usage.usageRange.range.start.line,
        matches: [
            {
                startLine: usage.usageRange.range.start.line,
                startCharacter: usage.usageRange.range.start.character,
                endLine: usage.usageRange.range.end.line,
                endCharacter: usage.usageRange.range.end.character,
            },
        ],
        plaintextLines: [usage.surroundingContent],
        highlightedHTMLRows: highlightedHTMLChunks?.[index],
        href: hrefForUsage(usage),
    }))
</script>

<div
    class="root"
    use:observeIntersection={scrollContainer ?? null}
    on:intersecting={event => (visible = visible || event.detail)}
>
    <div class="header">
        <CodeHostIcon {repository} />
        <span class="repo-name"><DisplayPath path={displayRepoName(repository)} /></span>
        <span class="interpunct">⋅</span>
        <span class="file-name">
            <DisplayPath
                {path}
                pathHref={pathHrefFactory({
                    repoName: repository,
                    revision: revision,
                    fullPath: path,
                    fullPathType: 'blob',
                })}
                showCopyButton
            />
        </span>
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
        flex-wrap: wrap;
        gap: 0.25rem 0.5rem;
        background-color: var(--secondary-4);
        padding: 0.375rem 0.75rem;
        border-bottom: 1px solid var(--border-color);
        align-items: center;

        position: sticky;
        top: 0;

        --icon-color: currentColor;
        :global([data-icon]) {
            flex: none;
        }

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

        .file-name {
            :global([data-path-container]) {
                flex-wrap: wrap;
            }
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
