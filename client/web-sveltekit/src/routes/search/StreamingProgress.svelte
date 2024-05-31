<script lang="ts">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'
    import ILucideCircleAlert from '~icons/lucide/circle-alert'
    import ILucideInfo from '~icons/lucide/info'
    import ILucideTriangleAlert from '~icons/lucide/triangle-alert'

    import { limitHit, sortBySeverity } from '$lib/branded'
    import { renderMarkdown, pluralize } from '$lib/common'
    import Icon2 from '$lib/Icon2.svelte'
    import Popover from '$lib/Popover.svelte'
    import ResultsIndicator from '$lib/search/resultsIndicator/ResultsIndicator.svelte'
    import SyntaxHighlightedQuery from '$lib/search/SyntaxHighlightedQuery.svelte'
    import type { Progress, Skipped } from '$lib/shared'
    import { Button } from '$lib/wildcard'

    export let progress: Progress
    export let state: 'complete' | 'error' | 'loading'

    const icons: Record<string, ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>> = {
        info: ILucideInfo,
        warning: ILucideTriangleAlert,
        error: ILucideCircleAlert,
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
    $: openItems = sortedItems.map((_, index) => index === 0)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: hasSuggestedItems = suggestedItems.length > 0
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: isError = severity === 'error' || state === 'error'
</script>

<Popover let:registerTrigger let:toggle placement="bottom-start">
    <Button variant={isError ? 'danger' : 'secondary'} size="sm" outline>
        <svelte:fragment slot="custom" let:buttonClass>
            <button use:registerTrigger class="{buttonClass} progress-button" on:click={() => toggle()}>
                <ResultsIndicator {state} {suggestedItems} {progress} {severity} />
            </button>
        </svelte:fragment>
    </Button>
    <div slot="content" class="streaming-popover">
        <p>
            Found {limitHit(progress) ? 'more than ' : ''}
            {progress.matchCount}
            {pluralize('result', progress.matchCount)}
            {#if progress.repositoriesCount !== undefined}
                from {progress.repositoriesCount} {pluralize('repository', progress.repositoriesCount, 'repositories')}.
            {/if}
        </p>
        {#if hasSkippedItems}
            <h3>Some results skipped</h3>
            {#each sortedItems as item, index (item.reason)}
                {@const open = openItems[index]}
                <button type="button" class="toggle" aria-expanded={open} on:click={() => (openItems[index] = !open)}>
                    <h4>
                        <Icon2 icon={icons[item.severity]} inline --icon-fill-color="var(--primary)" />
                        <span class="title">{item.title}</span>
                        {#if item.message}
                            <Icon2 icon={open ? ILucideChevronDown : ILucideChevronLeft} inline />
                        {/if}
                    </h4>
                </button>
                {#if item.message && open}
                    <div class="message">
                        {@html renderMarkdown(item.message)}
                    </div>
                {/if}
            {/each}
        {/if}
        {#if hasSuggestedItems}
            <p>Search again:</p>
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
                <Button variant="primary">
                    <svelte:fragment slot="custom" let:buttonClass>
                        <button class="{buttonClass} search" disabled={searchAgainDisabled}>
                            <Icon2 icon={ILucideSearch} />
                            <span>Search again</span>
                        </button>
                    </svelte:fragment>
                </Button>
            </form>
        {/if}
        <!--
        TODO: @jasonhawkharris - When we implement search jobs,
        we can change the link so that it points to where a user
        can actually create a search job
        -->
        {#if severity === 'error' || state === 'loading'}
            <div class="search-job-link">
                <small>
                    Search taking too long or timing out? Use <a
                        href="/help/code-search/types/search-jobs"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Search Job</a
                    > for background search.
                </small>
            </div>
        {/if}
    </div>
</Popover>

<style lang="scss">
    .chevron > :global(svg) {
        fill: currentColor !important;
    }

    .search-job-link {
        margin: 0rem 1rem 1rem 1rem;
        font-style: italic;
    }

    label {
        display: block;
    }

    .message {
        border-left: 2px solid var(--primary);
        padding-left: 0.5rem;
        margin: 0 1rem 1rem 1rem;
    }

    .progress-button {
        border: 1px solid var(--border-color-2);
        border-radius: 4px;
        padding: 0;
    }

    .streaming-popover {
        width: 24rem;

        p,
        h3,
        form {
            margin: 1rem;
        }
    }

    button.toggle {
        all: unset;

        cursor: pointer;
        display: block;
        box-sizing: border-box;
        padding: 0.5rem;
        width: 100%;

        h4 {
            display: flex;
            margin-bottom: 0;
            align-items: center;
            gap: 0.25rem;

            .title {
                flex: 1;
            }
        }
    }

    button.search {
        margin-top: 1rem;
    }
</style>
