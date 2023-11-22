<script lang="ts">
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import type { MatchGroup, ContentMatch } from '$lib/shared'

    import CodeExcerpt from './CodeExcerpt.svelte'

    export let result: ContentMatch
    export let grouped: MatchGroup[]

    $: ranges = grouped.map(group => ({
        startLine: group.startLine,
        endLine: group.endLine,
    }))

    async function fetchHighlightedFileMatchLineRanges(startLine: number, endLine: number) {
        const highlightedGroups = await fetchFileRangeMatches({ result, ranges })
        return highlightedGroups[grouped.findIndex(group => group.startLine === startLine && group.endLine === endLine)]
    }
</script>

<div class="root">
    {#each grouped as group}
        <div class="code">
            <CodeExcerpt
                startLine={group.startLine}
                endLine={group.endLine}
                fetchHighlightedFileRangeLines={async (...args) =>
                    group.blobLines ? group.blobLines : fetchHighlightedFileMatchLineRanges(...args)}
                matches={group.matches}
            />
        </div>
    {/each}
</div>

<style lang="scss">
    .root {
        border-radius: var(--border-radius);
        border: 1px solid var(--border-color);
        background-color: var(--code-bg);
    }

    .code {
        &:not(:first-child) {
            border-top: 1px solid var(--border-color);
        }
    }
</style>
