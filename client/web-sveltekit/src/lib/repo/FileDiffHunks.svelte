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

{#if hunks.length === 0}
    <div class="text-muted mr-2">No changes</div>
{:else}
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
                        @@ -{hunk.oldRange.startLine},{hunk.oldRange.lines} +{hunk.newRange.startLine},{hunk.newRange
                            .lines}
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
{/if}

<style lang="scss">
    table {
        width: 100%;
        border-collapse: collapse;
        font-family: var(--code-font-family);
        font-size: 0.75rem;
    }

    tr.added {
        --code-bg: var(--diff-add-bg);
    }

    tr.deleted {
        --code-bg: var(--diff-remove-bg);
    }

    td {
        background-color: var(--code-bg);

        &.num {
            min-width: 2.5rem;
            line-height: 1.6666666667;
            white-space: nowrap;
            text-align: right;
            -webkit-user-select: none;
            user-select: none;
            vertical-align: top;
            padding: 0 0.5rem;
            color: var(--text-muted);
        }

        &.header {
            white-space: pre-wrap;
            background-color: var(--color-bg-2);
            color: var(--body-color);
            padding: 0.25rem 1rem;
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
