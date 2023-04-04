<script lang="ts">
    import { createPopper, type Placement, type Options } from '@popperjs/core'
    import { afterUpdate } from 'svelte'

    export let tooltip: string
    export let placement: Placement = 'bottom'

    let visible = false
    let tooltipElement: HTMLElement
    let container: HTMLElement
    let target: Element | null
    let instance: ReturnType<typeof createPopper>

    function show() {
        visible = true
    }

    function hide() {
        visible = false
    }

    function updateInstance(options: Partial<Options>): void {
        if (instance) {
            instance.setOptions(options)
        }
    }

    afterUpdate(() => {
        instance?.update()
    })

    $: updateInstance({ placement })

    $: target = container?.firstElementChild
    $: if (tooltipElement && target && !instance) {
        instance = createPopper(target, tooltipElement, {
            placement,
            modifiers: [
                {
                    name: 'offset',
                    options: {
                        offset: [0, 8],
                    },
                },
            ],
        })
    }
</script>

<div
    class="target"
    bind:this={container}
    on:mouseenter={show}
    on:mouseleave={hide}
    on:focusin={show}
    on:focusout={hide}
>
    <slot />
</div>
<div class="tooltip-content" class:visible role="tooltip" bind:this={tooltipElement}>
    {tooltip}
    <div class="arrow" data-popper-arrow />
</div>

<style lang="scss">
    .target {
        display: contents;
    }

    .tooltip-content {
        --tooltip-font-size: 0.75rem; // 12px
        --tooltip-line-height: 1.02rem; // 16.32px / 16px, per Figma
        --tooltip-max-width: 256px;
        --tooltip-color: var(--light-text);
        --tooltip-border-radius: var(--border-radius);
        --tooltip-padding-y: 0.25rem;
        --tooltip-padding-x: 0.5rem;
        --tooltip-margin: 0;

        isolation: isolate;
        font-size: var(--tooltip-font-size);
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
        display: none;

        &:global([data-popper-placement^='top']) > .arrow {
            bottom: -4px;
        }

        &:global([data-popper-placement^='bottom']) > .arrow {
            top: -4px;
        }

        &:global([data-popper-placement^='left']) > .arrow {
            right: -4px;
        }

        &:global([data-popper-placement^='right']) > .arrow {
            left: -4px;
        }

        &.visible {
            display: block;
        }
    }

    .arrow,
    .arrow::before {
        position: absolute;
        width: 8px;
        height: 8px;
        background: inherit;
    }

    .arrow {
        visibility: hidden;

        &::before {
            visibility: visible;
            content: '';
            transform: rotate(45deg);
        }
    }
</style>
