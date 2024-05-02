<script lang="ts">
    import { numberWithCommas } from '$lib/common'
    import Timestamp from '$lib/Timestamp.svelte'
    import { Badge } from '$lib/wildcard'

    import type { GitReference_Ref } from './GitReference.gql'

    export let ref: GitReference_Ref

    $: authorName = ref.target.commit?.author.person.displayName ?? ''
    $: authorDate = ref.target.commit ? new Date(ref.target.commit.author.date) : null
    $: behind = ref.target.commit?.behindAhead?.behind
    $: ahead = ref.target.commit?.behindAhead?.ahead
</script>

<tr>
    <td>
        <Badge variant="link"><a href={ref.url}>{ref.displayName}</a></Badge>
    </td>
    <td colspan={behind || ahead ? 1 : 2}>
        <a href={ref.url}>
            <small
                >Updated {#if authorDate}<span><Timestamp date={authorDate} strict /></span>{/if}
                <span
                    >by
                    {authorName}</span
                ></small
            ></a
        >
    </td>
    {#if ahead || behind}
        <td class="diff">
            <a href={ref.url}
                ><small
                    ><span>{numberWithCommas(behind ?? 0)} behind,</span>
                    <span>{numberWithCommas(ahead ?? 0)} ahead</span></small
                ></a
            >
        </td>
    {/if}
</tr>

<style lang="scss">
    td {
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color-2);
    }

    tr:last-child td {
        border: none;
    }

    span {
        white-space: nowrap;
    }

    a {
        display: block;
    }

    small {
        color: var(--text-muted);

        &:hover {
            color: inherit;
        }
    }

    .diff {
        text-align: right;
    }
</style>
