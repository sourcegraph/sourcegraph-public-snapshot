<script lang="ts">
    import '$lib/highlight.scss'

    import range from 'lodash/range'

    import { highlightNodeMultiline } from '$lib/common'
    import { observeIntersection } from '$lib/intersection-observer'
    import type { MatchGroupMatch } from '$lib/shared'

    export let startLine: number
    export let endLine: number
    /**
     * Gets called when the code excerpt is visible in the viewport, to delay potentially expensive
     * highlighting operations until necessary.
     */
    export let fetchHighlightedFileRangeLines: (startLine: number, endLine: number) => Promise<string[]>
    export let matches: MatchGroupMatch[] = []

    function highlightMatches(node: HTMLElement, matches: MatchGroupMatch[]) {
        const visibleRows = node.querySelectorAll<HTMLTableRowElement>('tr')
        for (const highlight of matches) {
            // Select the HTML rows in the excerpt that correspond to the first and last line to be highlighted.
            // highlight.startLine is the 0-indexed line number in the code file, and startLine is the 0-indexed
            // line number of the first visible line in the excerpt. So, subtract startLine
            // from highlight.startLine to get the correct 0-based index in visibleRows that holds the HTML row
            // where highlighting should begin. Subtract startLine from highlight.endLine to get the correct 0-based
            // index in visibleRows that holds the HTML row where highlighting should end.
            const startRowIndex = highlight.startLine - startLine
            const endRowIndex = highlight.endLine - startLine
            const startRow = visibleRows[startRowIndex]
            const endRow = visibleRows[endRowIndex]
            if (startRow && endRow) {
                highlightNodeMultiline(
                    visibleRows,
                    startRow,
                    endRow,
                    startRowIndex,
                    endRowIndex,
                    highlight.startCharacter,
                    highlight.endCharacter
                )
            }
        }
    }

    let isVisible = false

    function onIntersection(event: { detail: boolean }) {
        // The component stays marked as "visible" if it was once to avoid
        // refetching highlighted lines and matches
        isVisible = isVisible || event.detail
    }
</script>

<code use:observeIntersection on:intersecting={onIntersection}>
    {#if isVisible}
        {#await fetchHighlightedFileRangeLines(startLine, endLine)}
            <!--create empty space to fill viewport to avoid layout shifts -->
            <table>
                <tbody>
                    {#each range(startLine, endLine) as index}
                        <tr>
                            <td class="line" data-line={index + 1} />
                            <td class="code" />
                        </tr>
                    {/each}
                </tbody>
            </table>
        {:then blobLines}
            {#key matches}
                <table use:highlightMatches={matches}>
                    {@html blobLines.join('')}
                </table>
            {/key}
        {/await}
    {/if}
</code>

<style lang="scss">
    code {
        display: flex;
        align-items: center;
        padding: 0.125rem 0.375rem;
        background-color: var(--background-color, --code-bg);
        overflow-x: auto;

        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
        line-height: 1rem;
        white-space: pre;

        :global(td.line::before) {
            content: attr(data-line);
            color: var(--text-muted);
        }

        :global(td.code) {
            white-space: pre;
            padding-left: 1rem;
        }
    }
</style>
