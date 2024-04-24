<script lang="ts" context="module">
    interface Match {
        startLine: number
        endLine: number
        startCharacter: number
        endCharacter: number
    }

    // Mulltiple locations will point to the same file, we only need to compute the
    // lines once.
    const lineCache = new Map<string, readonly string[]>()

    function getLines(location: ReferencePanelCodeExcerpt_Location): readonly string[] {
        const { resource } = location
        const key = `${resource.repository.id}:${resource.commit.oid}:${resource.path}`
        if (lineCache.has(key)) {
            return lineCache.get(key)!
        }
        const lines = resource.content.split(/\r?\n/)
        lineCache.set(key, lines)
        return lines
    }

    function adjustedLinesAndMatch(
        lines: readonly string[],
        range: NonNullable<ReferencePanelCodeExcerpt_Location['range']>
    ): [readonly string[], Match] {
        const trimmedLines = lines.map(line => line.trimStart())
        const match: Match = {
            startLine: range.start.line,
            endLine: range.end.line,
            startCharacter: range.start.character - (lines[0].length - trimmedLines[0].length),
            endCharacter:
                range.end.character - (lines[lines.length - 1].length - trimmedLines[trimmedLines.length - 1].length),
        }
        return [trimmedLines, match]
    }
</script>

<script lang="ts">
    import CodeExcerpt from '$lib/search/CodeExcerpt.svelte'
    import type { ReferencePanelCodeExcerpt_Location } from './ReferencePanelCodeExcerpt.gql'

    export let location: ReferencePanelCodeExcerpt_Location

    $: lines = location.range ? getLines(location).slice(location.range.start.line, location.range.end.line + 1) : []
    $: [plaintextLines, match] = location.range ? adjustedLinesAndMatch(lines, location.range) : [null, null]
</script>

{#if location.range && plaintextLines && match}
    <CodeExcerpt startLine={location.range.start.line} {plaintextLines} matches={[match]} />
{:else}
    <pre>(no content information)</pre>
{/if}

<style lang="scss">
</style>
