<script lang="ts">
    import { isErrorLike } from '$lib/common'
    import type { TooltipViewOptions } from '$lib/web'
    import type { Observable } from 'rxjs'
    import { readable } from 'svelte/store'
    import HovercardContent from './HovercardContent.svelte'
    import ActionItem from './ActionItem.svelte'

    type Observed<T> = T extends Observable<infer U> ? U : never

    export let hovercardData: TooltipViewOptions['hovercardData']

    $: hovercardDataStore = readable<Observed<TooltipViewOptions['hovercardData']> | null>(null, set => {
        const subscription = hovercardData.subscribe(set)
        return () => subscription.unsubscribe()
    })
    $: ({ actionsOrError, hoverOrError } = $hovercardDataStore ?? {})
    $: actions = !!actionsOrError && actionsOrError !== 'loading' && !isErrorLike(actionsOrError) ? actionsOrError : []
</script>

<div class="card cm-code-intel-hovercard">
    <div class="contents">
        {#if isErrorLike(hoverOrError)}
            <!-- TODO: Implement Alert component -->
            <div class="alert alert-danger">{hoverOrError.message}</div>
        {:else if hoverOrError === undefined || hoverOrError === 'loading'}
            <!-- do nothing -->
        {:else if hoverOrError === null || hoverOrError.contents.length === 0}
            <!-- Show some content to give the close button space and communicate to the user we couldn't find a hover. -->
            <small class="hover-empty">No hover information available.</small>
        {:else}
            {#each hoverOrError.contents as content}
                <HovercardContent {content} aggregatedBages={hoverOrError.aggregatedBadges} />
            {/each}
        {/if}
    </div>
    <div class="actions-container">
        <div class="actions">
            {#each actions as actionItem}
                <span class="action">
                    <ActionItem {actionItem} />
                </span>
            {/each}
        </div>
    </div>
</div>

<style lang="scss">
    .card {
        --hover-overlay-content-margin-top: 0.5rem;
        --hover-overlay-contents-right-padding: 1rem;
        --hover-overlay-horizontal-padding: 1rem;
        --hover-overlay-separator-color: var(--border-color);
        --hover-overlay-vertical-padding: 0.25rem;

        // Fixes the issue with `position: sticky` of the close button in Safari.
        // The sticky element misbehaves because `.card` has a `display: flex` rule.
        // Minimal example: https://codepen.io/valerybugakov/pen/ExWWOao?editors=1100
        display: block;
        min-width: 6rem;
        max-width: 34rem; // was 32rem; + 2rem to fit maximum code intel alert text
        z-index: 100;
        // Make sure content doesn't leak behind border-radius
        padding-bottom: var(--hover-overlay-vertical-padding);

        // From wildcard Card component
        --card-bg: var(--color-bg-1);
        --card-border-color: var(--border-color-2);
        --card-border-radius: var(--border-radius);
        --hover-box-shadow: 0 0 0 1px var(--primary) inset;
        position: relative;
        display: flex;
        flex-direction: column;
        min-width: 0;
        word-wrap: break-word;
        background-color: var(--card-bg);
        background-clip: border-box;
        border-width: 1px;
        border-style: solid;
        border-color: var(--card-border-color);
        border-radius: var(--card-border-radius) !important;

        // From WebHoverOverlay.module.scss
        --hover-overlay-content-color: var(--text-muted);
        --hover-overlay-separator-color: var(--border-color-2);

        border-color: var(--border-color);
        box-shadow: var(--dropdown-shadow);
        border-radius: var(--popover-border-radius);
    }

    .contents {
        overflow-x: hidden;
        overflow-y: auto;
        max-height: 10rem;
        padding-top: var(--hover-overlay-vertical-padding);
        padding-bottom: 0;
        padding-left: var(--hover-overlay-horizontal-padding);
        padding-right: var(--hover-overlay-contents-right-padding);
    }

    .actions-container {
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-top: 1px solid var(--hover-overlay-separator-color);
    }

    .actions {
        display: flex;
        align-items: center;
        padding-top: 0.75rem;
        padding-bottom: 0.5rem;
        padding-left: var(--hover-overlay-horizontal-padding);
        padding-right: var(--hover-overlay-horizontal-padding);

        span:not(:first-child) {
            margin-left: 0.5rem;
        }
    }
</style>
