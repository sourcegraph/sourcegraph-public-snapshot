<script lang="ts">
    import type { Placement } from '@floating-ui/dom'
    import type { Action } from 'svelte/action'

    import { popover, onClickOutside, portal } from './dom'

    export let placement: Placement = 'bottom'
    /**
     * Show the popover when hovering over the trigger.
     */
    export let showOnHover = false
    export let hoverDelay: number = 0
    export let hoverCloseDelay: number = 0

    let isOpen = false
    let trigger: HTMLElement | null
    let target: HTMLElement | undefined
    let popoverContainer: HTMLElement | null
    let delayTimer: ReturnType<typeof setTimeout>

    function toggle(open?: boolean): void {
        isOpen = open === undefined ? !isOpen : open
    }

    function handleClickOutside(event: { detail: HTMLElement }): void {
        if (event.detail !== trigger && !trigger?.contains(event.detail)) {
            isOpen = false
        }
    }

    const registerTarget: Action<HTMLElement> = node => {
        target = node
    }

    const registerTrigger: Action<HTMLElement> = node => {
        trigger = node

        function handleMouseEnterTrigger(): void {
            clearTimeout(delayTimer)
            delayTimer = setTimeout(() => {
                isOpen = true
            }, hoverDelay)
        }

        function handleMouseLeaveTrigger(event: MouseEvent): void {
            // It should be possible to move the mouse from the trigger to the popover without closing it
            if (event.relatedTarget && !popoverContainer?.contains(event.relatedTarget as Node)) {
                clearTimeout(delayTimer)
                delayTimer = setTimeout(() => {
                    isOpen = false
                }, hoverCloseDelay)
            }
        }

        function handleMouseMoveTrigger(): void {
            clearTimeout(delayTimer)
            delayTimer = setTimeout(() => {
                isOpen = true
            }, hoverDelay)
        }

        function handleWindowLoseFocus(): void {
            clearTimeout(delayTimer)
            isOpen = false
        }

        if (showOnHover) {
            node.addEventListener('mouseenter', handleMouseEnterTrigger)
            node.addEventListener('mouseleave', handleMouseLeaveTrigger)
            node.addEventListener('mousemove', handleMouseMoveTrigger)
            window.addEventListener('blur', handleWindowLoseFocus)
        }

        return {
            destroy() {
                trigger = null
                node.removeEventListener('mouseenter', handleMouseEnterTrigger)
                node.removeEventListener('mouseleave', handleMouseLeaveTrigger)
                node.removeEventListener('mousemove', handleMouseMoveTrigger)
                window.removeEventListener('blur', handleWindowLoseFocus)
            },
        }
    }

    const registerPopoverContainer: Action<HTMLElement> = node => {
        popoverContainer = node
        function handleMouseLeavePopover(event: MouseEvent): void {
            // It should be possible to move the mouse from the popover to the trigger without closing it
            if (event.relatedTarget && !trigger?.contains(event.relatedTarget as Node)) {
                delayTimer = setTimeout(() => toggle(false), hoverCloseDelay)
            }
        }

        function handleMouseEnterPopover(): void {
            // When the mouse enters the popover, cancel any pending close events
            clearTimeout(delayTimer)
        }

        if (showOnHover) {
            node.addEventListener('mouseenter', handleMouseEnterPopover)
            node.addEventListener('mouseleave', handleMouseLeavePopover)
        }
        return {
            destroy() {
                popoverContainer = null
                node.removeEventListener('mouseenter', handleMouseEnterPopover)
                node.removeEventListener('mouseleave', handleMouseLeavePopover)
            },
        }
    }
</script>

<slot {toggle} {registerTrigger} {registerTarget} />
{#if trigger && isOpen}
    <div
        use:portal
        use:onClickOutside
        use:registerPopoverContainer
        use:popover={{
            reference: target ?? trigger,
            options: {
                placement,
                offset: showOnHover ? 0 : 3,
                shift: { padding: 4 },
            },
        }}
        on:click-outside={handleClickOutside}
    >
        <slot name="content" {toggle} />
    </div>
{/if}

<style lang="scss">
    div {
        position: absolute;
        isolation: isolate;
        min-width: 10rem;
        font-size: 0.875rem;
        background-clip: padding-box;
        background-color: var(--dropdown-bg);
        color: var(--body-color);
        box-shadow: var(--popover-shadow);

        border: 1px solid var(--dropdown-border-color);
        border-radius: var(--popover-border-radius);
        // Ensure child elements do not overflow the border radius
        overflow: hidden;

        // We always display the popover on hover, but there may not be anything
        // inside until something we load something. This ensures we do not
        // render an empty border if there is nothing to show.
        &:empty {
            display: none;
        }
    }
</style>
