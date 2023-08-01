<script lang="ts">
    import { mdiAlertCircle, mdiChevronDown, mdiChevronLeft, mdiInformationOutline, mdiMagnify } from '@mdi/js'

    import { limitHit, sortBySeverity } from '$lib/branded'
    import { renderMarkdown, pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import SyntaxHighlightedQuery from '$lib/search/SyntaxHighlightedQuery.svelte'
    import type { Progress, Skipped } from '$lib/shared'
    import { Button } from '$lib/wildcard'

    export let progress: Progress

    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlertCircle,
        error: mdiAlertCircle,
    }
    let searchAgainDisabled = true

    function updateButton(event: Event) {
        const element = event.target as HTMLInputElement
        searchAgainDisabled = Array.from(element.form?.querySelectorAll('[name="query"]') ?? []).every(
            input => !(input as HTMLInputElement).checked
        )
    }

    $: matchCount = progress.matchCount + (progress.skipped.length > 0 ? '+' : '')
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: hasSkippedItems = progress.skipped.length > 0
    $: sortedItems = sortBySeverity(progress.skipped)
    $: openItems = sortedItems.map((_, index) => index === 0)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: hasSuggestedItems = suggestedItems.length > 0
</script>

<Popover let:registerTrigger let:toggle placement="bottom-start">
    <Button variant="secondary" size="sm" outline>
        <button slot="custom" let:className use:registerTrigger class={className} on:click={() => toggle()}>
            <Icon svgPath={icons[severity]} inline />
            {matchCount} results in {(progress.durationMs / 1000).toFixed(2)}s
            <Icon svgPath={mdiChevronDown} inline />
        </button>
    </Button>
    <div slot="content" class="streaming-popover">
        <p>
            Found {limitHit(progress) ? 'more than ' : ''}
            {progress.matchCount}
            {pluralize('result', progress.matchCount)}
            {#if progress.repositoriesCount !== undefined}
                from {progress.repositoriesCount} {pluralize('repository', progress.repositoriesCount, 'repositories')}
            {/if}.
        </p>
        {#if hasSkippedItems}
            <h3>Some results skipped</h3>
            {#each sortedItems as item, index (item.reason)}
                {@const open = openItems[index]}
                <Button variant="primary" outline>
                    <button
                        slot="custom"
                        type="button"
                        let:className
                        class="{className} p-2 w-100 bg-transparent border-0"
                        aria-expanded={open}
                        on:click={() => (openItems[index] = !open)}
                    >
                        <h4 class="d-flex align-items-center mb-0 w-100">
                            <span class="mr-1 flex-shrink-0"><Icon svgPath={icons[item.severity]} inline /></span>
                            <span class="flex-grow-1 text-left">{item.title}</span>
                            {#if item.message}
                                <span class="chevron flex-shrink-0"
                                    ><Icon svgPath={open ? mdiChevronDown : mdiChevronLeft} inline /></span
                                >
                            {/if}
                        </h4>
                    </button>
                </Button>
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
                    <button slot="custom" let:className class="{className} mt-3" disabled={searchAgainDisabled}>
                        <Icon svgPath={mdiMagnify} />
                        <span>Search again</span>
                    </button>
                </Button>
            </form>
        {/if}
    </div>
</Popover>

<style lang="scss">
    div.streaming-popover {
        width: 20rem;

        p,
        h3,
        form {
            margin: 1rem;
        }
    }

    .chevron > :global(svg) {
        fill: currentColor !important;
    }

    div.message {
        border-left: 2px solid var(--primary);
        padding-left: 0.5rem;
        margin: 0 1rem 1rem 1rem;
    }

    label {
        display: block;
    }
</style>
