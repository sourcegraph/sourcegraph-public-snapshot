<script lang="ts">
    import { capitalize } from 'lodash'

    import { sortBySeverity } from '$lib/branded'
    import type { Progress, Skipped } from '$lib/shared'

    export let progress: Progress
    export let suggestedItems: Required<Skipped>[] = []
    export let severity: string
    export let state: 'error' | 'complete' | 'loading'

    const SEE_MORE = 'See more details'
    const CENTER_DOT = '\u00B7' // AKA 'interpunct'

    interface ItemsBySeverity {
        items: Skipped[]
        mostSevere: Skipped | null
    }

    function filterAndSortItems(items: Skipped[]): ItemsBySeverity {
        if (items.length === 0) {
            return { items, mostSevere: null }
        }

        const filteredItems = items.filter((item: Skipped) => {
            return !['repository-fork', 'excluded-archive'].includes(item.reason)
        })

        const sorted = sortBySeverity(filteredItems)
        return { items: sorted, mostSevere: sorted[0] }
    }

    $: ({ items, mostSevere } = filterAndSortItems(progress.skipped))
    $: suggestedItems = items.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: isError = severity === 'error' || state === 'error'
    $: hasSkippedItems = progress.skipped.length > 0
</script>

<div class="suggested-action">
    {#if hasSkippedItems && mostSevere}
        <small class="info-badge" class:error-text={isError}>{capitalize(mostSevere?.title ?? mostSevere.title)}</small>
        {#if mostSevere.suggested}
            <small>{CENTER_DOT}</small>
            <small>{capitalize(mostSevere?.suggested ? mostSevere.suggested.title : '')}</small>
            <small class="code-font">{mostSevere.suggested?.queryExpression}</small>
        {/if}
    {:else}
        <small>{SEE_MORE}</small>
    {/if}
</div>

<style lang="scss">
    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: flex-end;
        column-gap: 0.5rem;
        color: var(--text-muted);
        white-space: nowrap;
        font-size: var(--font-size-xs);

        .info-badge {
            background-color: var(--primary-2);
            border-radius: 3px;
            padding: 0rem 0.15rem;
            color: var(--text-title);
            white-space: nowrap;

            &.error-text {
                background: var(--danger-2);
            }
        }

        .code-font {
            background-color: var(--secondary);
            border-radius: 3px;
            padding: 0rem 0.15rem;
            color: var(--text-title);
            white-space: nowrap;
            font-family: var(--code-font-family);
        }
    }
</style>
