<script lang="ts">
    import { createPopper, type Placement } from '@popperjs/core'
    import { onClickOutside } from './dom'

    export let placement: Placement = 'bottom'

    let isOpen = false
    let trigger: HTMLElement | null
    let content: HTMLElement | null

    function createInstance(target: HTMLElement, content: HTMLElement) {
        return createPopper(target, content, {
            placement,
        })
    }

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

    $: if (isOpen && trigger && content) {
        createInstance(trigger, content)
    }
</script>

<slot {toggle} {registerTrigger} />
{#if isOpen}
    <div class="content" bind:this={content} use:onClickOutside on:click-outside={clickOutside}>
        <slot name="content" />
    </div>
{/if}

<style lang="scss">
    .content {
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
