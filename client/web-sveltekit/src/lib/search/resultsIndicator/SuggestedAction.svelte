<script lang="ts">
    import { capitalize } from 'lodash'

    import { sortBySeverity } from '$lib/branded'
    import type { Progress, Skipped } from '$lib/shared'

    export let progress: Progress
    export let suggestedItems: Required<Skipped>[] = []
    export let severity: string
    export let state: 'error' | 'complete' | 'loading'

    const CENTER_DOT = '\u00B7' // AKA 'interpunct'

    $: sortedItems = sortBySeverity(progress.skipped)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: isError = severity === 'error' || state === 'error'
    $: hasSkippedItems = progress.skipped.length > 0
    $: mostSevere = sortedItems[0]
    $: done = progress.done
</script>

<div class="action-container" class:error-text={isError}>
    <div class="suggested-action">
        <!-- completed search -->
        {#if done && !hasSkippedItems}
            <div class="more-details">
                <small> See more details </small>
            </div>
        {/if}

        <!-- completed with skipped items -->
        {#if done && hasSkippedItems}
            <div class="info-badge" class:error-text={isError}>
                <small>
                    {capitalize(mostSevere?.title ?? mostSevere.title)}
                </small>
            </div>
        {/if}

        <!-- completed with suggested items -->
        {#if done && mostSevere && mostSevere.suggested}
            <div class="separator">{CENTER_DOT}</div>
            <div class="action-badge">
                <small>
                    {capitalize(mostSevere?.suggested ? mostSevere.suggested.title : '')}&nbsp;
                    <span class="code-font">
                        {mostSevere.suggested?.queryExpression}
                    </span>
                </small>
            </div>
        {/if}
        <!--
        TODO: @jasonhawkharris - When we implement search jobs,
        we can change the link so that it points to where a user
        can actually create a search job. We should also change
        the text of the link when we do so, "Create a search job"
        -->
        {#if severity === 'error' && !mostSevere.suggested}
            <div class="separator">{CENTER_DOT}</div>
            <div class="search-job-link">
                <small>
                    Use <a href="/help/code-search/types/search-jobs">Search Job</a> for background search.
                </small>
            </div>
        {/if}
    </div>
</div>

<style lang="scss">
    .code-font {
        background-color: var(--secondary);
        border-radius: 3px;
        color: var(--text-body);
        font-family: var(--code-font-family);
        padding: 0rem 0.2rem;
    }

    .info-badge {
        background-color: var(--primary-2);
        border-radius: 3px;
        color: var(--text-body);
        padding: 0rem 0.2rem 0rem 0.2rem;

        &.error-text {
            background: var(--danger-2);
        }
    }

    .more-details {
        color: var(--gray-06);
    }

    .separator {
        padding-left: 0.4rem;
        padding-right: 0.4rem;
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
    }
    .search-job-link {
        // we don't want the text of this message to change color
        // on hover when in an error state, so we set the text color explicitly
        color: var(--text-body);
    }
</style>
