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

    $: sortedItems = sortBySeverity(progress.skipped)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: isError = severity === 'error' || state === 'error'
    $: hasSkippedItems = progress.skipped.length > 0
    $: mostSevere = sortedItems[0]
    $: done = progress.done
    $: forkedSuggestion = mostSevere.reason === 'excluded-fork' || mostSevere.reason === 'repository-fork'
    $: archivedSuggestion = mostSevere.reason === 'excluded-archive' || mostSevere.reason === 'repository-archive'
    $: forkedOrArchived = forkedSuggestion || archivedSuggestion
</script>

<div class="suggested-action">
    <!-- completed search with no skipped items -->
    {#if done && !hasSkippedItems}
        <div class="more-details"><small>{SEE_MORE}</small></div>
    {/if}

    <!-- completed with skipped items -->
    {#if done && hasSkippedItems && !forkedOrArchived}
        <div class="info-badge" class:error-text={isError}>
            <small>{capitalize(mostSevere?.title ?? mostSevere.title)}</small>
        </div>
    {:else}
        <div class="more-details"><small>{SEE_MORE}</small></div>
    {/if}

    <!-- completed with suggested items -->
    {#if done && mostSevere && mostSevere.suggested && !forkedOrArchived}
        <small class="separator">{CENTER_DOT}</small>
        <small class="action-badge">
            {capitalize(mostSevere?.suggested ? mostSevere.suggested.title : '')}&nbsp;
            <span class="code-font">{mostSevere.suggested?.queryExpression}</span>
        </small>
    {/if}

    <!--
    TODO: @jasonhawkharris - When we implement search jobs,
    we can change the link so that it points to where a user
    can actually create a search job. We should also change
    the text of the link when we do so, "Create a search job"
    -->
    {#if severity === 'error' && !mostSevere.suggested}
        <div class="error">
            <small>{CENTER_DOT}</small>
            <small>
                Use <a href="/help/code-search/types/search-jobs">Search Job</a> for background search.
            </small>
        </div>
    {/if}
</div>

<style lang="scss">
    .info-badge {
        background-color: var(--primary-2);
        border-radius: 3px;
        padding: 0rem 0.2rem 0rem 0.2rem;

        &.error-text {
            background: var(--danger-2);
        }
    }

    .more-details {
        color: var(--text-body);
    }

    .error {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: flex-end;
        gap: 0.5rem 0.25rem;
    }

    .separator {
        padding-left: 0.4rem;
        padding-right: 0.4rem;
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: space-evenly;
    }

    .code-font {
        background-color: var(--secondary);
        border-radius: 3px;
        font-family: var(--code-font-family);
        padding: 0rem 0.2rem;
    }
</style>
