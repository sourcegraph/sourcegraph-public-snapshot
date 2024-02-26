<script lang="ts">
    import { mdiAlertCircle, mdiChevronDown, mdiChevronLeft, mdiInformationOutline, mdiMagnify } from '@mdi/js'

    import { capitalizeFirstLetter } from '@sourcegraph/branded'

    import { getProgressText, limitHit, sortBySeverity } from '$lib/branded'
    import { renderMarkdown, pluralize } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Popover from '$lib/Popover.svelte'
    import SyntaxHighlightedQuery from '$lib/search/SyntaxHighlightedQuery.svelte'
    import type { Progress, Skipped } from '$lib/shared'
    import { Button } from '$lib/wildcard'

    import SecondCounter from './SecondCounter.svelte'

    const CENTER_DOT = '\u00B7' // interpunct
    export let progress: Progress
    export let loading: boolean

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

    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: hasSkippedItems = progress.skipped.length > 0
    $: sortedItems = sortBySeverity(progress.skipped)
    $: mostSevere = sortedItems[0]
    $: openItems = sortedItems.map((_, index) => index === 0)
    $: suggestedItems = sortedItems.filter((skipped): skipped is Required<Skipped> => !!skipped.suggested)
    $: hasSuggestedItems = suggestedItems.length > 0
</script>

<Popover let:registerTrigger let:toggle placement="bottom-start">
    <Button variant="secondary" size="sm" outline>
        <svelte:fragment slot="custom" let:buttonClass>
            <button use:registerTrigger class="{buttonClass} progress-button" on:click={() => toggle()}>
                {#if loading}
                    <div class="loading">
                        <LoadingSpinner inline />
                        <div class="messages">
                            <div class="progress-info-message">
                                Fetching results...
                                <SecondCounter />
                            </div>
                            <div class="loading-action">500+ results</div>
                        </div>
                    </div>
                {/if}
                {#if progress}
                    <div class="progress">
                        <Icon svgPath={icons[severity]} inline />
                        <div class="messages">
                            <div class="progress-info-message">
                                {getProgressText(progress).visibleText}
                            </div>
                            <div class="progress-action-message">
                                {#if hasSkippedItems}
                                    <div class={`info-badge ${mostSevere.severity === 'error' && 'error'}`}>
                                        {capitalizeFirstLetter(mostSevere.title)}
                                    </div>
                                {/if}
                                {#if hasSkippedItems && hasSuggestedItems}
                                    <div class="separator">{CENTER_DOT}</div>
                                {/if}
                                {#if hasSuggestedItems}
                                    <div class="action-badge">
                                        {capitalizeFirstLetter(mostSevere.suggested ? mostSevere.suggested.title : '')}
                                        <span class="code-font">
                                            {mostSevere.suggested?.queryExpression}
                                        </span>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </div>
                {/if}
                <Icon svgPath={mdiChevronDown} inline />
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
                <Button variant="primary" outline>
                    <svelte:fragment slot="custom" let:buttonClass>
                        <button
                            type="button"
                            class="{buttonClass} p-2 w-100 bg-transparent border-0"
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
                    </svelte:fragment>
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
                    <svelte:fragment slot="custom" let:buttonClass>
                        <button class="{buttonClass} mt-3" disabled={searchAgainDisabled}>
                            <Icon svgPath={mdiMagnify} />
                            <span>Search again</span>
                        </button>
                    </svelte:fragment>
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

    .error {
        background-color: var(--danger);
    }

    .progress-button {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        justify-items: flex-start;
        margin-bottom: 1rem;
        margin-top: 1rem;
    }

    div.progress {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        line-height: 1.2;
        margin-right: 1rem;
    }

    div.messages {
        align-content: center;
        align-items: flex-start;
        display: flex;
        flex-flow: column nowrap;
        margin-left: 0.5rem;
    }

    div.progress-info-message {
        font-size: 0.9rem;
    }

    div.separator {
        margin-left: 0.4rem;
        margin-right: 0.4rem;
        padding-top: 0.1rem;
    }

    div.progress-action-message {
        display: flex;
        flex-flow: row nowrap;
        font-family: var(--base-font-family);
        font-size: 0.7rem;
        justify-content: space-around;
        margin-top: 0.3rem;
    }

    .loading-action {
        color: var(--text-muted);
        display: flex;
        flex-flow: row nowrap;
        font-family: var(--base-font-family);
        font-size: 0.7rem;
        justify-content: space-around;
    }

    div.info-badge {
        background: var(--primary);
        border-radius: 3px;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
        padding-top: 0.1rem;
    }

    div.info-badge.error {
        background: var(--danger);
    }

    div.action-badge {
        padding-top: 0.1rem;
    }

    .code-font {
        font-family: var(--code-font-family);
        background-color: var(--gray-08);
        border-radius: 3px;
        padding-left: 0.1rem;
        padding-right: 0.1rem;
    }

    .loading {
        display: flex;
        flex-flow: row nowrap;
        margin-right: 1rem;
    }
</style>
