<script lang="ts">
    import '$lib/highlight.scss'

    import range from 'lodash/range'

    import { highlightNodeMultiline } from '$lib/common'
    import { observeIntersection } from '$lib/intersection-observer'
    import type { MatchGroupMatch } from '$lib/shared'

    export let startLine: number
    export let endLine: number
    export let plaintextLines: string[]
    export let highlightedHTMLRows: string[] | undefined = undefined

    /**
     * Gets called when the code excerpt is visible in the viewport, to delay potentially expensive
     * highlighting operations until necessary.
     */
    export let matches: MatchGroupMatch[] = []
</script>

<code>
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
