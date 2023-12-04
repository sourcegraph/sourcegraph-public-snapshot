<script lang="ts">
    import type { Placement } from '@popperjs/core'

    import { createPopover, onClickOutside } from './dom'
    import { afterUpdate } from 'svelte'

    export let placement: Placement = 'bottom'

    const { update, popover } = createPopover()

    let isOpen = false
    let trigger: HTMLElement | null

    function toggle(open?: boolean): void {
        isOpen = open === undefined ? !isOpen : open
    }

    function clickOutside(event: { detail: HTMLElement }): void {
        if (event.detail !== trigger && !trigger?.contains(event.detail)) {
            isOpen = false
        }
    }

    function registerTrigger(node: HTMLElement) {
        trigger = node
    }

    afterUpdate(update)
</script>

<slot {toggle} {registerTrigger} />
{#if trigger && isOpen}
    <div use:popover={{ target: trigger, options: { placement } }} use:onClickOutside on:click-outside={clickOutside}>
        <slot name="content" {toggle} />
    </div>
{/if}

<style lang="scss">
    div {
        isolation: isolate;
        z-index: 1000;
        min-width: 10rem;
        font-size: 0.875rem;
        background-clip: padding-box;
        background-color: var(--dropdown-bg);
        border: 1px solid var(--dropdown-border-color);
        border-radius: var(--popover-border-radius);
        color: var(--body-color);
        box-shadow: var(--dropdown-shadow);
    }
</style>
