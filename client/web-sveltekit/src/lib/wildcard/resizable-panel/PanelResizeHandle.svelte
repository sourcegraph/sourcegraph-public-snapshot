<script context="module" lang="ts">
    export type ResizeHandlerState = 'drag' | 'hover' | 'inactive'
</script>

<script lang="ts">
    import { getContext, onMount } from 'svelte'
    import { writable, type Writable } from 'svelte/store'
    import classNames from 'classnames'
    import { getId } from './utils/common'
    import { assert } from './utils/assert'
    import type { PanelGroupContext, ResizeHandlerAction, ResizeEvent } from './types'
    import { PanelResizeHandleRegistry } from '$lib/wildcard/resizable-panel/PanelResizeHandleRegistry'

    export let id: string | null = null
    export let className: string = ''

    const resizeHandleId = id ?? getId()
    const { direction, groupId, startDragging, stopDragging, getResizeHandler } =
        getContext<PanelGroupContext>('panel-group-context')

    // Local state
    let element: HTMLElement
    const stateStore: Writable<ResizeHandlerState> = writable('inactive')

    onMount(() => {
        const resizeHandler = getResizeHandler(resizeHandleId)

        assert(element, 'Element ref not attached')

        const setResizeHandlerState = (action: ResizeHandlerAction, isActive: boolean, event: ResizeEvent) => {
            if (isActive) {
                switch (action) {
                    case 'down': {
                        stateStore.set('drag')
                        startDragging(resizeHandleId, event)
                        break
                    }
                    case 'move': {
                        if (state !== 'drag') {
                            stateStore.set('hover')
                        }

                        resizeHandler(event)
                        break
                    }
                    case 'up': {
                        stateStore.set('hover')
                        stopDragging()
                    }
                }
            } else {
                stateStore.set('inactive')
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

    $: state = $stateStore
</script>

<div
    class={classNames(className, 'separator')}
    role="separator"
    tabIndex="0"
    bind:this={element}
    data-resize-handle
    data-panel-group-id={groupId}
    data-panel-group-direction={direction}
    data-resize-handle-state={state}
    data-resize-handle-active={state === 'drag' ? 'pointer' :  undefined}
    data-panel-resize-handle-id={resizeHandleId}
>
    <slot />
</div>

<style lang="scss">
    .separator {
        touch-action: none;
        user-select: none;
    }
</style>
