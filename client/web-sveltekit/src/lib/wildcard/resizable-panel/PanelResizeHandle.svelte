<script context="module" lang="ts">
    export type ResizeHandlerState = 'drag' | 'hover' | 'inactive'
</script>

<script lang="ts">
    import { getContext, onMount } from 'svelte'
    import { writable, type Writable } from 'svelte/store'

    import { PanelResizeHandleRegistry } from '$lib/wildcard/resizable-panel/PanelResizeHandleRegistry'

    import type { PanelGroupContext, ResizeHandlerAction, ResizeEvent } from './types'
    import { assert } from './utils/assert'
    import { getId } from './utils/common'

    export let id: string | null = null

    const resizeHandleId = id ?? getId()
    const { direction, groupId, startDragging, stopDragging, getResizeHandler } =
        getContext<PanelGroupContext>('panel-group-context')

    // Local state
    let element: HTMLElement
    const state: Writable<ResizeHandlerState> = writable('inactive')

    onMount(() => {
        const resizeHandler = getResizeHandler(resizeHandleId)

        assert(element, 'Element ref not attached')

        const setResizeHandlerState = (action: ResizeHandlerAction, isActive: boolean, event: ResizeEvent) => {
            if (isActive) {
                switch (action) {
                    case 'down': {
                        $state = 'drag'
                        startDragging(resizeHandleId, event)
                        break
                    }
                    case 'move': {
                        if ($state !== 'drag') {
                            $state = 'hover'
                        }

                        resizeHandler(event)
                        break
                    }
                    case 'up': {
                        $state = 'hover'
                        stopDragging()
                    }
                }
            } else {
                $state = 'inactive'
            }
        }

        return PanelResizeHandleRegistry.registerResizeHandle(
            resizeHandleId,
            element,
            direction,
            { coarse: 15, fine: 5 },
            setResizeHandlerState
        )
    })

    // TODO [VK]: Add keyboard handlers aka WindowSplitterResizeHandlerBehavior
    // https://www.w3.org/WAI/ARIA/apg/patterns/windowsplitter/
</script>

<div
    class="separator"
    role="separator"
    tabIndex="0"
    bind:this={element}
    data-resize-handle
    data-panel-group-id={groupId}
    data-panel-group-direction={direction}
    data-resize-handle-state={$state}
    data-resize-handle-active={$state === 'drag' ? 'pointer' : undefined}
    data-panel-resize-handle-id={resizeHandleId}
/>

<style lang="scss">
    $resize-handle-bg: var(--border-color);
    $resize-handle-hover-bg: var(--border-color-2);
    $resize-handle-drag-bg: var(--oc-blue-3);
    $resize-handle-size: 1px;
    $resize-handle-active-size: 3px;

    .separator {
        // Since drag handler is always rendered within flex
        // PanelGroup component is safe to assume that flex rules
        // can applied here.
        flex: 0 0 $resize-handle-size;
        display: flex;
        touch-action: none;
        user-select: none;
        width: 100%;
        height: 100%;
        background: $resize-handle-bg;
        position: relative;

        &::before {
            content: '';
            z-index: 1;
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            width: 100%;
            height: 100%;
            min-width: $resize-handle-active-size;
            min-height: $resize-handle-active-size;
        }

        &[data-resize-handle-state='hover']::before {
            display: block;
            background: $resize-handle-hover-bg;
        }

        &[data-resize-handle-state='drag']::before {
            display: block;
            background: $resize-handle-drag-bg;
        }
    }
</style>
