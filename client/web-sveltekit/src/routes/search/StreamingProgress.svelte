<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import { goto } from '$app/navigation'
    import { limitHit, sortBySeverity } from '$lib/branded'
    import { renderMarkdown, pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Popover from '$lib/Popover.svelte'
    import ResultsIndicator from '$lib/search/resultsIndicator/ResultsIndicator.svelte'
    import SyntaxHighlightedQuery from '$lib/search/SyntaxHighlightedQuery.svelte'
    import type { Progress, Skipped } from '$lib/shared'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import { Alert, Button } from '$lib/wildcard'
    import ProductStatusBadge from '$lib/wildcard/ProductStatusBadge.svelte'

    import type { SearchJob } from './searchJob'

    export let progress: Progress
    export let state: 'complete' | 'error' | 'loading'
    export let searchJob: SearchJob | undefined

    const icons: Record<string, ComponentProps<Icon>['icon']> = {
        info: ILucideInfo,
        warning: ILucideAlertCircle,
        error: ILucideCircleX,
    }
    let searchAgainDisabled = true

    function updateButton(event: Event) {
        const element = event.target as HTMLInputElement
        searchAgainDisabled = Array.from(element.form?.querySelectorAll('[name="query"]') ?? []).every(
            input => !(input as HTMLInputElement).checked
        )
    }

    $: hasSkippedItems = progress.skipped.length > 0
    $: sortedItems = sortBySeverity(progress.skipped)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: hasSuggestedItems = suggestedItems.length > 0
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: isError = severity === 'error' || state === 'error'

    $: if (searchJob && !$searchJob?.validating) {
        TELEMETRY_RECORDER.recordEvent('search.exhaustiveJobs', 'view', {
            metadata: { validState: $searchJob?.validationError ? 0 : 1 },
        })
    }
</script>

<Popover let:registerTrigger let:toggle placement="bottom-start" flip={false}>
    <Button variant={isError ? 'danger' : 'secondary'} size="sm" outline>
        <svelte:fragment slot="custom" let:buttonClass>
            <button
                use:registerTrigger
                class="{buttonClass} progress-button"
                on:click={() => toggle()}
                data-testid="page.search-results.progress-button"
            >
                <ResultsIndicator {state} {suggestedItems} {progress} {severity} />
            </button>
        </svelte:fragment>
    </Button>
    <div slot="content" class="streaming-popover" data-testid="page.search-results.progress-popover">
        <div class="section">
            Found {limitHit(progress) ? 'more than ' : ''}
            {progress.matchCount}
            {pluralize('result', progress.matchCount)}
            {#if progress.repositoriesCount !== undefined}
                from {progress.repositoriesCount} {pluralize('repository', progress.repositoriesCount, 'repositories')}.
            {/if}
            {#if hasSkippedItems}
                <details open={sortedItems.length === 0}>
                    <summary>
                        <Icon icon={ILucideInfo} aria-hidden inline --icon-color="var(--primary)" />
                        Why was the limit reached?
                    </summary>

                    <ul class="skipped-items">
                        {#each sortedItems as item (item.reason)}
                            <li>
                                {#if item.message}
                                    <details open={sortedItems.length === 0}>
                                        <summary>
                                            <Icon
                                                icon={icons[item.severity]}
                                                aria-label={item.severity}
                                                inline
                                                --icon-color="var(--primary)"
                                            />
                                            <span class="title">{item.title}</span>
                                        </summary>
                                        <div class="message">
                                            {@html renderMarkdown(item.message)}
                                        </div>
                                    </details>
                                {:else}
                                    <Icon
                                        icon={icons[item.severity]}
                                        aria-label={item.severity}
                                        inline
                                        --icon-color="var(--primary)"
                                    />
                                    <span class="title">{item.title}</span>
                                {/if}
                            </li>
                        {/each}
                    </ul></details
                >
            {/if}
        </div>
        {#if hasSuggestedItems}
            <div class="section">
                <h4>Search again:</h4>
                <form on:submit|preventDefault>
                    {#each suggestedItems as item (item.suggested.queryExpression)}
                        <label>
                            <input
                                type="checkbox"
                                name="query"
                                value={item.suggested.queryExpression}
                                on:click={updateButton}
                            />
                            {item.suggested.title} (
                            <SyntaxHighlightedQuery query={item.suggested.queryExpression} />
                            )
                        </label>
                    {/each}
                    <Button variant="primary" size="sm">
                        <svelte:fragment slot="custom" let:buttonClass>
                            <button class="{buttonClass} search" disabled={searchAgainDisabled}>
                                <Icon icon={ILucideSearch} aria-hidden="true" inline />
                                Modify and rerun
                            </button>
                        </svelte:fragment>
                    </Button>
                </form>
            </div>
        {/if}
        {#if searchJob}
            <div class="section">
                <h4>Create a search job: <ProductStatusBadge status="beta" /></h4>
                {#if $searchJob?.validationError}
                    <Alert variant="info">
                        {$searchJob.validationError.message}
                    </Alert>
                {/if}
                <p
                    >Search jobs exhaustively return all matches of a query. Results can be downloaded in JSON Lines
                    format.</p
                >
                {#if $searchJob?.creationError}
                    <Alert variant="danger">
                        {$searchJob.creationError.message}
                    </Alert>
                {/if}
                <Button
                    variant="secondary"
                    size="sm"
                    disabled={!!$searchJob?.validationError || $searchJob?.validating || $searchJob?.creating}
                    on:click={() =>
                        searchJob?.create().then(
                            () => {
                                TELEMETRY_RECORDER.recordEvent('search.exhaustiveJobs.create', 'click')
                                goto('/search-jobs')
                            },
                            () => {
                                /* do nothing */
                            }
                        )}
                >
                    {#if $searchJob?.creating}
                        <LoadingSpinner inline />
                        <span class="loading">Starting search job</span>
                    {:else}
                        <Icon icon={ILucideSearch} aria-hidden="true" inline />
                        Create a search job
                    {/if}
                </Button>
            </div>
        {/if}
    </div>
</Popover>

<style lang="scss">
    .streaming-popover {
        max-width: 24rem;

        > .section > details {
            margin-top: 0.5rem;
        }
    }

    .section {
        padding: 1rem;
        border-top: 1px solid var(--border-color-2);

        &:first-child {
            border-top: none;
        }
    }

    label {
        display: block;
        // Resets global style
        font-weight: normal;
        margin-left: 1rem;
    }

    h4 {
        font-weight: normal;
    }

    .message {
        border-left: 2px solid var(--primary);
        padding-left: 1rem;
        margin: 0.5rem 0 0 1.125rem;
    }

    .title {
        padding-left: 0.25rem;
    }

    .progress-button {
        border: 1px solid var(--border-color-2);
        border-radius: 4px;
        padding: 0;
    }

    button.search {
        margin-top: 1rem;
    }

    ul.skipped-items {
        list-style: none;
        margin: 0;
        padding: 0 0 0 1rem;

        li {
            margin-top: 0.25rem;
            padding: 0.25rem 0;

            & + li {
                border-top: 1px solid var(--border-color-2);
            }
        }
    }

    summary {
        cursor: pointer;
    }

    span.loading {
        vertical-align: middle;
    }
</style>
