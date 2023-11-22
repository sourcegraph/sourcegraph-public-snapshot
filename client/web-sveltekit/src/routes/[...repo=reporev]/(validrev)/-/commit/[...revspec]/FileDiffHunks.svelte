<script lang="ts">
    import '$lib/highlight.scss'

    import { DiffHunkLineType, type FileDiffFields } from '$lib/graphql-operations'

    export let hunks: FileDiffFields['hunks']
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
                <tr>
                    <td colspan="2" />
                    <td class="content">
                        @@ -{hunk.oldRange.startLine},{hunk.oldRange.lines} +{hunk.newRange.startLine},{hunk.newRange
                            .lines}
                        {#if hunk.section}
                            @@ {hunk.section}
                        {/if}
                    </td>
                </tr>
                {#each hunk.highlight.lines as line, i (line)}
                    {@const both = line.kind === DiffHunkLineType.UNCHANGED}
                    {@const addition = line.kind === DiffHunkLineType.ADDED}
                    {@const deletion = line.kind === DiffHunkLineType.DELETED}
                    {@const marker = addition ? '+' : deletion ? '-' : ' '}

                    <tr class:both class:addition class:deletion>
                        <td class="num"
                            >{#if !addition}{hunk.oldRange.startLine - 1 + i}{/if}</td
                        >
                        <td class="num"
                            >{#if !deletion}{hunk.newRange.startLine - 1 + i}{/if}</td
                        >
                        <td class="content" data-diff-marker={marker}>{@html line.html}</td>
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
        border-radius: var(--border-radius);
        border: 1px solid var(--border-color);
    }

    tr {
        font-family: var(--code-font-family);
        font-size: 0.75rem;
    }

    tr.addition {
        --code-bg: var(--diff-add-bg);
    }

    tr.deletion {
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
            vertical-align: top !important;
            padding: 0 0.5rem;
        }

        &.content {
            padding-left: 0.5rem;
            padding-right: 0.5rem;
            white-space: pre-wrap;
            color: var(--body-color);

            & > :global(*) {
                display: inline-block;
            }

            &::before {
                padding-right: 0.5rem;
                content: attr(data-diff-marker);
            }
        }
    }
</style>
