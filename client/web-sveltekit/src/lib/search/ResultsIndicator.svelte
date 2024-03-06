<script lang="ts">
    import { mdiChevronDown, mdiInformationOutline, mdiAlert, mdiAlertCircle } from '@mdi/js'

    import { getProgressText } from '$lib/branded'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { capitalizeFirstLetter } from '$lib/search/utils'
    import type { Progress, Skipped } from '$lib/shared'
    import { Button } from '$lib/wildcard'

    export let hasSkippedItems: boolean
    export let sortedItems: Skipped[]
    export let hasSuggestedItems: boolean
    export let searchProgress: Progress
    export let togglePopover: () => void
    export let trigger: (node: HTMLElement) => void

    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlert,
        error: mdiAlertCircle,
    }
    const CENTER_DOT = '\u00B7' // interpunct

    $: loading = searchProgress.skipped.length === 0
    $: mostSevere = sortedItems[0]
    $: severity = searchProgress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
</script>

<Button variant="secondary" size="sm" outline>
    <svelte:fragment slot="custom" let:buttonClass>
        <button use:trigger class="{buttonClass} progress-button" on:click={() => togglePopover()}>
            {#if loading && !searchProgress}
                <div class="loading">
                    <LoadingSpinner inline />
                    <div class="messages">
                        <div class="progress-info-message">
                            Fetching results...
                            <!--<SecondCounter />-->
                        </div>
                        <div class="loading-action">500+ results</div>
                    </div>
                </div>
            {/if}
            {#if searchProgress}
                <div class="progress">
                    <Icon svgPath={icons[severity]} size={18} />
                    <div class="messages">
                        <div class="progress-info-message">
                            {getProgressText(searchProgress).visibleText}
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
            <Icon svgPath={mdiChevronDown} size={18} />
        </button>
    </svelte:fragment>
</Button>

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

    div.progress {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        line-height: 1.2;
        margin-right: 1rem;
        background-color: transparent;
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
        margin-right: 0.4rem;
        margin-left: 0.4rem;
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
        padding-left: 0.2rem;
        padding-right: 0.2rem;
        padding-top: 0.1rem;
        padding-bottom: 0.1rem;
    }

    .loading {
        display: flex;
        flex-flow: row nowrap;
        margin-right: 1rem;
    }
</style>
