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
            const isNotFork = item.reason !== 'repository-fork'
            const isNotArchive = item.reason !== 'excluded-archive'
            return isNotFork && isNotArchive
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
        <div class="info-badge" class:error-text={isError}>
            <small>{capitalize(mostSevere?.title ?? mostSevere.title)}</small>
        </div>
        {#if mostSevere.suggested}
            <small class="separator">{CENTER_DOT}</small>
            <small class="action-badge">
                {capitalize(mostSevere?.suggested ? mostSevere.suggested.title : '')}&nbsp;
                <span class="code-font">{mostSevere.suggested?.queryExpression}</span>
            </small>
        {/if}
    {:else}
        <div class="more-details"><small>{SEE_MORE}</small></div>
    {/if}
</div>

<style lang="scss">
    .info-badge {
        background-color: var(--primary-2);
        border-radius: 3px;
        padding-left: 0.25rem;
        padding-right: 0.25rem;

        &.error-text {
            background: var(--danger-2);
        }
    }

    .more-details {
        color: var(--text-muted);
    }

    .separator {
        padding-left: 0.4rem;
        padding-right: 0.4rem;
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: flex-end;
    }

    .code-font {
        background-color: var(--secondary);
        border-radius: 3px;
        font-family: var(--code-font-family);
        padding-right: 0.25rem;
        padding-left: 0.25rem;
        padding-top: 0.25rem;
        padding-bottom: 0.25rem;
    }
</style>
