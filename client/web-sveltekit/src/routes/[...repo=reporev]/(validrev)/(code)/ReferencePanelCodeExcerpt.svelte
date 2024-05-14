<script lang="ts" context="module">
    // Multiple locations will point to the same file, we only need to compute the
    // lines once.
    const lineCache = new Map<string, readonly string[]>()

    function getLines(resource: ReferencePanelCodeExcerpt_Location['resource']): readonly string[] {
        const key = `${resource.repository.id}:${resource.commit.oid}:${resource.path}`
        if (lineCache.has(key)) {
            return lineCache.get(key)!
        }
        const lines = resource.content.split(/\r?\n/)
        lineCache.set(key, lines)
        return lines
    }
</script>

<script lang="ts">
    import { derived, readable } from 'svelte/store'
    import { observeIntersection } from '$lib/intersection-observer'
    import CodeExcerpt from '$lib/CodeExcerpt.svelte'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import { toReadable } from '$lib/utils'

    import type { ReferencePanelCodeExcerpt_Location } from './ReferencePanelCodeExcerpt.gql'

    export let location: ReferencePanelCodeExcerpt_Location

    $: plaintextLines = location.range
        ? getLines(location.resource).slice(location.range.start.line, location.range.end.line + 1)
        : []
    $: matches = location.range
        ? [
              {
                  startLine: location.range.start.line,
                  endLine: location.range.end.line,
                  startCharacter: location.range.start.character,
                  endCharacter: location.range.end.character,
              },
          ]
        : []

    let visible = false
    // We rely on fetchFileRangeMatches to cache the result for us so that repeated
    // calls will not result in repeated network requests.
    $: highlightedHTMLRows =
        visible && location.range
            ? derived(
                  toReadable(
                      fetchFileRangeMatches({
                          result: {
                              repository: location.resource.repository.name,
                              commit: location.resource.commit.oid,
                              path: location.resource.path,
                          },
                          ranges: [{ startLine: location.range.start.line, endLine: location.range.end.line + 1 }],
                      })
                  ),
                  result => result.value?.[0] || []
              )
            : readable([])
</script>

{#if location.range && plaintextLines.length > 0}
    <div use:observeIntersection on:intersecting={event => (visible = visible || event.detail)}>
        <CodeExcerpt
            collapseWhitespace
            hideLineNumbers
            startLine={location.range.start.line}
            {plaintextLines}
            {matches}
            highlightedHTMLRows={$highlightedHTMLRows}
        />
    </div>
{:else}
    <pre>(no content information)</pre>
{/if}
