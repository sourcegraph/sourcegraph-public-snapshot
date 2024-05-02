<svelte:options immutable />

<script lang="ts">
    import '$lib/highlight.scss'

    import { highlightNodeMultiline } from '$lib/common'
    import type { MatchGroupMatch } from '$lib/shared'

    /**
     * Number of the first line in the code excerpt. This is 1-indexed.
     * Doesn't have any effect when `highlightedHTMLRows` or `hideLineNumbers` are set.
     */
    export let startLine: number
    export let plaintextLines: readonly string[]
    export let highlightedHTMLRows: readonly string[] | undefined = undefined
    export let matches: MatchGroupMatch[] = []
    /**
     * Causes whitespace to *not* be preserved. Can be useful to ignore the leading whitespace in a code block,
     * but will also remove any intentional whitespace formatting.
     */
    export let collapseWhitespace = false
    export let hideLineNumbers = false

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
</script>

<code class:collapseWhitespace class:hideLineNumbers>
    {#key matches}
        {#if highlightedHTMLRows === undefined || highlightedHTMLRows.length === 0}
            <table use:highlightMatches={matches}>
                <tbody>
                    {#each plaintextLines as line, index}
                        <tr>
                            <td class="line" data-line={startLine + index + 1} />
                            <td class="code">{line}</td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {:else}
            <table use:highlightMatches={matches}>
                <tbody>
                    {@html highlightedHTMLRows.join('')}
                </tbody>
            </table>
        {/if}
    {/key}
</code>

<style lang="scss">
    code {
        display: flex;
        align-items: center;
        background-color: var(--background-color, --code-bg);
        overflow-x: auto;

        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
        line-height: 1.1rem;
        white-space: pre;

        :global(td.line::before) {
            content: attr(data-line);
            color: var(--text-muted);
            padding-right: 1rem;
        }

        :global(td.code) {
            white-space: inherit;
        }

        &.collapseWhitespace {
            white-space: normal;
        }

        &.hideLineNumbers {
            :global(td.line::before) {
                display: none;
            }
        }
    }
</style>
