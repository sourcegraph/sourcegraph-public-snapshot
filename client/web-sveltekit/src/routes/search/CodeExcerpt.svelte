<script lang="ts">
    import '@sourcegraph/wildcard/src/global-styles/highlight.scss'

    import range from 'lodash/range'
    import { of, type Observable } from 'rxjs'
    import { catchError } from 'rxjs/operators'

    import { asError, isErrorLike, highlightNodeMultiline } from '$lib/common'
    import type { MatchGroupMatch } from '$lib/shared'
    import { observeIntersection } from '$lib/intersection-observer'

    export let startLine: number
    export let endLine: number
    export let fetchHighlightedFileRangeLines: (startLine: number, endLine: number) => Observable<string[]>
    export let blobLines: string[] | undefined = undefined
    export let matches: MatchGroupMatch[] = []

    let blobLinesOrError: string[] | Error | undefined = undefined

    function highlightRanges(node: HTMLElement, matches: MatchGroupMatch[]) {
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
        isVisible = event.detail
    }

    $: if (isVisible) {
        const observable = blobLines ? of(blobLines) : fetchHighlightedFileRangeLines(startLine, endLine)
        observable.pipe(catchError(error => [asError(error)])).subscribe(_blobLinesOrError => {
            blobLinesOrError = _blobLinesOrError
        })
    }
</script>

<code use:observeIntersection on:intersecting={onIntersection}>
    {#if blobLinesOrError && !isErrorLike(blobLinesOrError)}
        {#key blobLinesOrError}
            <table use:highlightRanges={matches}>
                {@html blobLinesOrError.join('')}
            </table>
        {/key}
    {:else if !blobLinesOrError}
        <table>
            <tbody>
                {#each range(startLine, endLine) as index}
                    <tr>
                        <td class="line" data-line={index + 1} />
                        <!--create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) -->
                        <td class="code" />
                    </tr>
                {/each}
            </tbody>
        </table>
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
