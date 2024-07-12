<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import { pathHrefFactory } from '$lib/path/index'
    import { getRevision, type ContentMatch, type PathMatch, type SymbolMatch } from '$lib/shared'

    import RepoRev from './RepoRev.svelte'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: rev = result.branches?.[0]

    $: matches =
        result.type !== 'symbol' && result.pathMatches
            ? result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
            : []
</script>

<div class="root">
    <RepoRev repoName={result.repository} {rev} />
    <span class="interpunct">·</span>
    <span class="path" use:highlightRanges={{ ranges: matches }}>
        <DisplayPath
            path={result.path}
            pathHref={pathHrefFactory({
                repoName: result.repository,
                revision: getRevision(result.branches, result.commit),
                fullPath: result.path,
                fullPathType: 'blob',
            })}
            showCopyButton
        />
    </span>
</div>

<style lang="scss">
    .root {
        display: inline-flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;

        :global([data-path-container]) {
            flex-flow: wrap;
        }
    }

    .path {
        display: contents;
    }

    .interpunct {
        color: var(--text-disabled);
    }
</style>
