<script lang="ts">
    import { DiffHunkLineType } from '$lib/graphql-types'

    import '$lib/highlight.scss'

    import type { FileDiffHunks_Hunk } from './FileDiffHunks.gql'

    export let hunks: readonly FileDiffHunks_Hunk[]

    function linesToDiffInformation(lines: FileDiffHunks_Hunk['highlight']['lines']): {
        newLineOffset: number
        oldLineOffset: number
        marker: '-' | '+' | ' '
        added: boolean
        deleted: boolean
        html: string
    }[] {
        let newLineOffset = -1
        let oldLineOffset = -1

        return lines.map(line => {
            let marker: '-' | '+' | ' ' = ' '
            let added = false
            let deleted = false

            switch (line.kind) {
                case DiffHunkLineType.ADDED:
                    marker = '+'
                    newLineOffset++
                    added = true
                    break
                case DiffHunkLineType.DELETED:
                    marker = '-'
                    oldLineOffset++
                    deleted = true
                    break
                case DiffHunkLineType.UNCHANGED:
                    newLineOffset++
                    oldLineOffset++
                    break
            }

            return { newLineOffset, oldLineOffset, marker, added, deleted, html: line.html }
        })
    }
</script>

<table>
    <colgroup>
        <col width="40" />
        <col width="40" />
        <col />
    </colgroup>
    <tbody>
        {#each hunks as hunk (hunk.oldRange.startLine)}
            {@const oldStartLine = hunk.oldRange.startLine}
            {@const newStartLine = hunk.newRange.startLine}
            <tr>
                <td class="header" colspan="3">
                    @@ -{hunk.oldRange.startLine},{hunk.oldRange.lines} +{hunk.newRange.startLine},{hunk.newRange.lines}
                    {#if hunk.section}
                        @@ {hunk.section}
                    {/if}
                </td>
            </tr>
            {#each linesToDiffInformation(hunk.highlight.lines) as { marker, added, deleted, newLineOffset, oldLineOffset, html }}
                <tr class:added class:deleted>
                    <td class="num"
                        >{#if !added}{oldStartLine + oldLineOffset}{/if}</td
                    >
                    <td class="num"
                        >{#if !deleted}{newStartLine + newLineOffset}{/if}</td
                    >
                    <td class="content" data-diff-marker={marker}>{@html html}</td>
                </tr>
            {/each}
        {/each}
    </tbody>
</table>

<style lang="scss">
    table {
        width: 100%;
        border-collapse: collapse;
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
    }

    tr.added {
        --code-bg: var(--diff-add-bg);
    }

    tr.deleted {
        --code-bg: var(--diff-remove-bg);
    }

    td {
        background-color: var(--code-bg);
        line-height: var(--code-line-height);

        &.header {
            white-space: pre-wrap;
            background-color: var(--color-bg-2);
            color: var(--body-color);
            padding: 0.25rem 1rem;
        }

        &.num {
            color: var(--text-muted);
            min-width: 2.5rem;
            // The alignment between the line numbers and the marker/content seems to be a bit off, due to,
            // apparently, specifying `vertical-align: top`. Hover this declaration is necessary so that the
            // line number is aligned with the first line when the content wraps around.
            // Adding a top padding makes it look better.
            padding-top: 2px;
            padding-inline: 0.5rem;
            text-align: right;
            user-select: none;
            vertical-align: top;
            white-space: nowrap;
        }

        &.content {
            white-space: pre-wrap;

            > :global(*) {
                display: inline;
            }

            &::before {
                padding-right: 0.5rem;
                content: attr(data-diff-marker);
                color: var(--text-muted);
            }
        }
    }
</style>
