<script lang="ts">
    import { numberWithCommas } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import DiffSquares from '$lib/repo/DiffSquares.svelte'
    import { Badge } from '$lib/wildcard'

    import type { DiffSummaryHeaderCommit } from './DiffSummaryHeader.gql'

    export let commit: DiffSummaryHeaderCommit

    $: parent = commit.ancestors.nodes.at(1)
    $: diffStat = commit.diff.fileDiffs.diffStat
</script>

<span class="root">
    <span class="label">
        <Badge variant="link">
            <a href={commit.canonicalURL}>{commit.abbreviatedOID}</a>
        </Badge>&nbsp;(selected)</span
    >
    <Icon icon={ILucideGitCompareArrows} inline aria-hidden />
    <span class="label">
        {#if parent}
            <Badge variant="link">
                <a href={parent.canonicalURL}>{parent.abbreviatedOID}</a>
            </Badge>&nbsp;(parent)
        {:else}
            (parent unavailable)
        {/if}
    </span>
    <span class="squares">
        <small>
            {#if diffStat.added > 0}
                <span class="added">+{numberWithCommas(diffStat.added)}</span>
            {/if}
            {#if diffStat.deleted > 0}
                <span class="deleted">-{numberWithCommas(diffStat.deleted)}</span>
            {/if}
        </small>
        &nbsp;
        <DiffSquares added={diffStat.added} deleted={diffStat.deleted} />
    </span>
</span>

<style lang="scss">
    .root {
        display: inline-grid;
        grid-template-columns: auto auto auto auto auto;
        grid-template-areas: 'x x x . squares';
        grid-template-rows: auto;

        align-items: center;
        gap: 1rem;

        @media (--mobile) {
            gap: 0.25rem;
            grid-template-columns: auto auto auto;
            grid-template-areas: 'x x x' 'squares squares squares';
        }
    }

    small,
    .squares,
    .label {
        white-space: nowrap;
    }

    .label {
        color: var(--text-muted);
    }

    .squares {
        grid-area: squares;
        display: inline-flex;
        align-items: center;
        font-weight: 700;

        .added {
            color: var(--success);
        }

        .deleted {
            color: var(--danger);
        }
    }
</style>
