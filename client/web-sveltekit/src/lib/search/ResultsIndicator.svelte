<script lang="ts">
    import { mdiChevronDown, mdiInformationOutline, mdiAlert, mdiAlertCircle } from '@mdi/js'
    import { onMount } from 'svelte'

    import { getProgressText } from '$lib/branded'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { capitalizeFirstLetter } from '$lib/search/utils'
    import type { Progress, Skipped } from '$lib/shared'

    export let hasSkippedItems: boolean
    export let sortedItems: Skipped[]
    export let hasSuggestedItems: boolean
    export let searchProgress: Progress

    let elapsedSeconds: number = 0
    let elapsedMilliseconds: number = 0
    let displaySeconds: number = 0
    let displayMilliseconds: number = 0

    let lastUpdate = Date.now()
    onMount(() => {
        const startTime = Date.now()
        setInterval(() => {
            const now = Date.now()
            const delta = now - startTime
            elapsedSeconds = Math.floor(delta / 1000)
            elapsedMilliseconds = delta % 100
            if (now - lastUpdate >= 5) {
                displaySeconds = elapsedSeconds
                displayMilliseconds = elapsedMilliseconds
                lastUpdate = now
            }
        }, 1300)
    })

    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlert,
        error: mdiAlertCircle,
    }
    const CENTER_DOT = '\u00B7' // interpunct
    const MAX_SEARCH_DURATION = 15

    $: loading = searchProgress.skipped.length === 0
    $: mostSevere = sortedItems[0]
    $: severity = searchProgress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
</script>

{#if !loading && searchProgress}
    <div class="loading">
        <LoadingSpinner inline />
        <div class="messages">
            <div class="progress-info-message">
                Fetching results... {elapsedSeconds}.{elapsedMilliseconds}s
            </div>
            {#if elapsedSeconds < MAX_SEARCH_DURATION}
                <div class="loading-action-message">Running search...</div>
            {:else}
                <div class="duration-badge">
                    <div class={`info-badge duration`}>Taking too long?</div>
                    <div class="separator">{CENTER_DOT}</div>
                    <div class="action-badge">
                        Use
                        <a href="https://sourcegraph.com/docs/code-search/types/search-jobs" target="_blank">
                            Search Job
                        </a>
                        for background search
                    </div>
                </div>
            {/if}
        </div>
    </div>
{/if}
{#if !searchProgress}
    <div class="progress">
        <Icon svgPath={icons[severity]} size={18} />
        <div class="messages">
            <div class="progress-info-message">
                {getProgressText(searchProgress).visibleText}
            </div>
            <div class="progress-action-message">
                {#if !hasSkippedItems}
                    <div class="more-details">See more details</div>
                {:else}
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
<Icon svgPath={mdiChevronDown} size={18} />

<style lang="scss">
    .error {
        background-color: var(--danger);
    }

    .progress-button {
        border: none;
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        justify-items: flex-start;
    }

    .progress {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        line-height: 1.2;
        margin-right: 1rem;
        background-color: transparent;
    }

    .messages {
        align-content: center;
        align-items: flex-start;
        display: flex;
        flex-flow: column nowrap;
        margin-left: 0.5rem;
    }

    .progress-info-message {
        font-size: 0.9rem;
    }

    .separator {
        margin-right: 0.4rem;
        margin-left: 0.4rem;
    }

    .progress-action-message {
        display: flex;
        flex-flow: row nowrap;
        font-family: var(--base-font-family);
        font-size: 0.7rem;
        justify-content: space-around;
        margin-top: 0.5rem;
    }

    .loading-action {
        color: var(--text-muted);
        display: flex;
        flex-flow: row nowrap;
        font-family: var(--base-font-family);
        font-size: 0.7rem;
        justify-content: space-around;
    }

    .info-badge {
        background: var(--primary);
        border-radius: 3px;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .info-badge.error {
        background: var(--danger);
    }

    .info-badge.duration {
        background: var(--warning);
        color: black;
    }

    .code-font {
        font-family: var(--code-font-family);
        background-color: var(--gray-08);
        border-radius: 3px;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .loading {
        display: flex;
        flex-flow: row nowrap;
        margin-right: 1rem;
    }

    .duration-badge {
        display: flex;
        flex-flow: row-nowrap;
        margin-top: 0.5rem;
    }

    .loading-action-message {
        margin-top: 0.5rem;
        color: var(--gray-06);
    }

    .more-details {
        color: var(--gray-06);
    }
</style>
