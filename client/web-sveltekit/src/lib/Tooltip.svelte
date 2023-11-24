<script lang="ts" context="module">
    import type { Placement } from '@popperjs/core'
    import { placements } from '@popperjs/core'

    export type { Placement }
    export { placements }
</script>

<script lang="ts">
    import { createPopover, uniqueID } from './dom'
    import { afterUpdate } from 'svelte'

    /**
     * The content of the tooltip.
     */
    export let tooltip: string
    /**
     * On which side to show the tooltip by default.
     */
    export let placement: Placement = 'bottom'
    /**
     * Force the tooltip to be always visible
     * (only used for stories).
     */
    export let alwaysVisible = false

    const id = uniqueID('tooltip')
    const { update, popover } = createPopover()

    let visible = false
    let container: HTMLElement
    let target: Element | null

    afterUpdate(update)

    function show() {
        visible = true
    }

    function hide() {
        visible = false
    }

    $: options = {
        placement,
        modifiers: [
            {
                name: 'offset',
                options: {
                    offset: [0, 8],
                },
            },
        ],
    }
    $: target = container?.firstElementChild
    $: if (target) {
        target.setAttribute('aria-labeledby', id)
    }
</script>

<!-- TODO: close tooltip on escape -->
<!--
    These event handlers listen for bubbled events from the trigger. The element
    itself is not interactable.
    svelte-ignore a11y-no-static-element-interactions
-->
<div
    class="container"
    bind:this={container}
    on:mouseenter={show}
    on:mouseleave={hide}
    on:focusin={show}
    on:focusout={hide}
>
    <slot />
</div>
{#if (alwaysVisible || visible) && target && tooltip}
    <div role="tooltip" {id} use:popover={{ target, options }}>
        {tooltip}
        <div data-popper-arrow />
    </div>
{/if}

<style lang="scss">
    .container {
        display: contents;
    }

    [role='tooltip'] {
        --tooltip-font-size: 0.75rem; // 12px
        --tooltip-line-height: 1.02rem; // 16.32px / 16px, per Figma
        --tooltip-max-width: 256px;
        --tooltip-color: var(--light-text);
        --tooltip-border-radius: var(--border-radius);
        --tooltip-padding-y: 0.25rem;
        --tooltip-padding-x: 0.5rem;
        --tooltip-margin: 0;

        all: initial;
        isolation: isolate;
        font-family: inherit;
        font-size: var(--tooltip-font-size);
        font-style: normal;
        font-weight: normal;
        line-height: var(--tooltip-line-height);
        max-width: var(--tooltip-max-width);
        background-color: var(--tooltip-bg);
        border-radius: var(--tooltip-border-radius);
        color: var(--tooltip-color);
        padding: var(--tooltip-padding-y) var(--tooltip-padding-x);
        user-select: text;
        word-wrap: break-word;
        border: none;
        min-width: 0;
        z-index: 100;
    }

    :global([data-popper-placement^='top']) > [data-popper-arrow] {
        bottom: -4px;
    }

    :global([data-popper-placement^='bottom']) > [data-popper-arrow] {
        top: -4px;
    }

    :global([data-popper-placement^='left']) > [data-popper-arrow] {
        right: -4px;
    }

    :global([data-popper-placement^='right']) > [data-popper-arrow] {
        left: -4px;
    }

    [data-popper-arrow],
    [data-popper-arrow]::before {
        position: absolute;
        width: 8px;
        height: 8px;
        background: inherit;
    }

    [data-popper-arrow] {
        visibility: hidden;

        &::before {
            visibility: visible;
            content: '';
            transform: rotate(45deg);
        }
    }
</style>
