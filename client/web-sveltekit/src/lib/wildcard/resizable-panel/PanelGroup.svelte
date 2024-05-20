<script context="module" lang="ts">
    import { getId } from './utils/common'
    import type { PanelInfo, PanelsLayout } from './types'

    // Store save layout debounced callback globally
    // to limit the frequency of localStorage updates.
    const DEBOUNCE_MAP: Record<string, typeof savePanelGroupLayout> = {}
    const LOCAL_STORAGE_DEBOUNCE_INTERVAL = 100

    function getPanelGroupId(propId: string): string {
        return `resizable-group-panel-${propId ?? getId()}`
    }
</script>

<script lang="ts">
    // This is a forked ported to svelte version of react-resizable-panels package
    // It's licenced under MIT license, original source - https://github.com/bvaughn/react-resizable-panels/tree/main

    import { onDestroy, setContext } from 'svelte'
    import { derived, type Readable, type Unsubscriber, writable, type Writable } from 'svelte/store'

    import { Exceed, PanelResizeHandleRegistry } from './PanelResizeHandleRegistry'
    import { assert } from './utils/assert'
    import { debounce } from './utils/debounce'
    import { loadPanelGroupLayout, savePanelGroupLayout } from './utils/storage'
    import { fuzzyNumbersEqual } from './numbers'
    import { findPanelDataIndex, getPanelMetadata, sortPanels } from './utils/common'
    import { calculateUnsafeDefaultLayout } from './utils/calculateUnsafeDefaultLayout'
    import { validatePanelGroupLayout } from './utils/validatePanelGroupLayout'
    import { computePanelFlexBoxStyle } from './utils/computePanelFlexBoxStyle'
    import { determinePivotIndices } from './utils/determinePivotIndices'
    import { calculateDeltaPercentage } from './utils/calculateDeltaPercentage'
    import { adjustLayoutByDelta } from './utils/adjustLayoutByDelta'
    import { compareLayouts } from './utils/compareLayouts'
    import { getResizeHandleElement } from './utils/dom'
    import { getResizeEventCursorPosition } from './event/getResizeEventCursorPosition'
    import { isKeyDown, isMouseEvent, isTouchEvent } from './event'
    import type { Direction, DragState, PanelGroupContext, PanelId, ResizeEvent } from './types'
    import { PanelGroupDirection } from './types'
    import { callPanelCallbacks } from './utils/callPanelsCallbacks'

    // Props
    export let id: string = ''
    export let direction: Direction = PanelGroupDirection.Horizontal

    // Local state
    let groupElement: HTMLElement
    let panelSizeBeforeCollapseMap = new Map<string, number>()
    let panelIdToLastNotifiedSizeMap: Record<string, number> = {}

    // Used in resize handler to avoid cursor panel UI flickering
    let prevDelta: number | null = null
    const groupId = getPanelGroupId(id)

    const layoutStore: Writable<PanelsLayout> = writable([])
    const panelsStore: Writable<PanelInfo[]> = writable([])
    const dragStateStore: Writable<DragState | null> = writable(null)

    onDestroy(
        // Connect registered panels store with layout store
        // set initial layout as we populate new panel elements
        panelsStore.subscribe(panels => {
            let unsafeLayout: PanelsLayout | null = null

            if (id && $layoutStore.length > 0) {
                const state = loadPanelGroupLayout(id, panels)

                if (state) {
                    unsafeLayout = state.layout
                    panelSizeBeforeCollapseMap = new Map(Object.entries(state.expandToSizes))
                }
            }

            if (unsafeLayout === null) {
                unsafeLayout = calculateUnsafeDefaultLayout(panels)
            }

            // Validate even saved layouts in case something has changed since last render
            // e.g. for pixel groups, this could be the size of the window
            const nextLayout = validatePanelGroupLayout(
                unsafeLayout,
                panels.map(panelData => panelData.constraints)
            )

            $layoutStore = nextLayout

            callPanelCallbacks($panelsStore, nextLayout, panelIdToLastNotifiedSizeMap)
        })
    )

    onDestroy(
        layoutStore.subscribe(layout => {
            if (!id || layout.length === 0 || layout.length !== $panelsStore.length) {
                return
            }

            let debouncedSave = DEBOUNCE_MAP[id]

            if (debouncedSave == null) {
                debouncedSave = debounce(savePanelGroupLayout, LOCAL_STORAGE_DEBOUNCE_INTERVAL)

                DEBOUNCE_MAP[id] = debouncedSave
            }

            // Clone mutable data before passing to the debounced
            // function, else we run the risk of saving an incorrect
            // combination of mutable and immutable values to state.
            debouncedSave(id, [...$panelsStore], layout, new Map(panelSizeBeforeCollapseMap))
        })
    )

    // External API methods
    function collapsePanel(panel: PanelInfo): void {
        if (!panel.constraints.collapsible) {
            return
        }

        const panelConstraintsArray = $panelsStore.map(panel => panel.constraints)
        const { collapsedSize = 0, panelSize, pivotIndices } = getPanelMetadata($panelsStore, panel, $layoutStore)

        assert(panelSize != null, `Panel size not found for panel "${panel.id}"`)

        if (!fuzzyNumbersEqual(panelSize, collapsedSize)) {
            // Store size before collapse;
            // This is the size that gets restored if the expand() API is used.
            panelSizeBeforeCollapseMap.set(panel.id, panelSize)

            const isLastPanel = findPanelDataIndex($panelsStore, panel.id) === $panelsStore.length - 1
            const delta = isLastPanel ? panelSize - collapsedSize : collapsedSize - panelSize

            const nextLayout = adjustLayoutByDelta({
                delta,
                pivotIndices,
                initialLayout: $layoutStore,
                prevLayout: $layoutStore,
                panelConstraints: panelConstraintsArray,
                trigger: 'imperative-api',
            })

            if (!compareLayouts($layoutStore, nextLayout)) {
                $layoutStore = nextLayout

                callPanelCallbacks($panelsStore, nextLayout, panelIdToLastNotifiedSizeMap)
            }
        }
    }

    function expandPanel(panel: PanelInfo, minSizeOverride?: number): void {
        if (!panel.constraints.collapsible) {
            return
        }

        const {
            pivotIndices,
            panelSize = 0,
            collapsedSize = 0,
            minSize: minSizeFromProps = 0,
        } = getPanelMetadata($panelsStore, panel, $layoutStore)
        const panelConstraintsArray = $panelsStore.map(panel => panel.constraints)

        const minSize = minSizeOverride ?? minSizeFromProps

        if (fuzzyNumbersEqual(panelSize, collapsedSize)) {
            // Restore this panel to the size it was before it was collapsed, if possible.
            const prevPanelSize = panelSizeBeforeCollapseMap.get(panel.id)

            const baseSize = prevPanelSize != null && prevPanelSize >= minSize ? prevPanelSize : minSize
            const isLastPanel = findPanelDataIndex($panelsStore, panel.id) === $panelsStore.length - 1
            const delta = isLastPanel ? panelSize - baseSize : baseSize - panelSize

            const nextLayout = adjustLayoutByDelta({
                delta,
                initialLayout: $layoutStore,
                panelConstraints: panelConstraintsArray,
                pivotIndices,
                prevLayout: $layoutStore,
                trigger: 'imperative-api',
            })

            if (!compareLayouts($layoutStore, nextLayout)) {
                $layoutStore = nextLayout

                callPanelCallbacks($panelsStore, nextLayout, panelIdToLastNotifiedSizeMap)
            }
        }
    }

    function registerPanel(panel: PanelInfo): Unsubscriber {
        panelsStore.update(panels => {
            if (panels.some(existingPanel => existingPanel.id === panel.id)) {
                return panels
            }

            // Since ordering plays a big role when it comes to hide/show
            // and restore panel size we should keep panels list ordered
            // If order field isn't provided we rely on Svelte rendering order
            return sortPanels([...panels, panel])
        })

        return () => {
            // When a panel is removed from the group, we should delete the most
            // recent prev-size entry for it. If we don't do this, then a conditionally
            // rendered panel might not call onResize when it's re-mounted.
            delete panelIdToLastNotifiedSizeMap[panel.id]

            panelsStore.update(panels => panels.filter(existingTab => existingTab.id !== panel.id))
        }
    }

    function getResizeHandler(dragHandleId: string) {
        return function resizeHandler(event: ResizeEvent): void {
            event.preventDefault()

            if (!groupElement) {
                return
            }

            const { initialLayout } = $dragStateStore ?? {}
            const pivotIndices = determinePivotIndices(groupId, dragHandleId, groupElement)
            const delta = calculateDeltaPercentage(
                event,
                dragHandleId,
                direction as PanelGroupDirection,
                $dragStateStore,
                null,
                groupElement
            )

            if (delta === 0) {
                return
            }

            const isHorizontal = direction === 'horizontal'
            const panelConstraints = $panelsStore.map(panel => panel.constraints)

            const nextLayout = adjustLayoutByDelta({
                delta,
                initialLayout: initialLayout ?? $layoutStore,
                panelConstraints,
                pivotIndices,
                prevLayout: $layoutStore,
                trigger: isKeyDown(event) ? 'keyboard' : 'mouse-or-touch',
            })

            const layoutChanged = !compareLayouts($layoutStore, nextLayout)

            // Only update the cursor for layout changes triggered by touch/mouse events (not keyboard)
            // Update the cursor even if the layout hasn't changed (we may need to show an invalid cursor state)
            if (isMouseEvent(event) || isTouchEvent(event)) {
                // Watch for multiple subsequent deltas; this might occur for tiny cursor movements.
                // In this case, Panel sizes might not change–
                // but updating cursor in this scenario would cause a flicker.
                if (prevDelta !== delta) {
                    prevDelta = delta

                    if (!layoutChanged) {
                        // If the pointer has moved too far to resize the panel any further, note
                        // this so we can update the cursor. This mimics VS Code behavior.
                        if (isHorizontal) {
                            PanelResizeHandleRegistry.reportConstraintsViolation(
                                dragHandleId,
                                delta < 0 ? Exceed.HORIZONTAL_MIN : Exceed.HORIZONTAL_MAX
                            )
                        } else {
                            PanelResizeHandleRegistry.reportConstraintsViolation(
                                dragHandleId,
                                delta < 0 ? Exceed.VERTICAL_MIN : Exceed.VERTICAL_MAX
                            )
                        }
                    } else {
                        PanelResizeHandleRegistry.reportConstraintsViolation(dragHandleId, Exceed.NO_CONSTRAINT)
                    }
                }
            }

            if (layoutChanged) {
                $layoutStore = nextLayout

                callPanelCallbacks($panelsStore, nextLayout, panelIdToLastNotifiedSizeMap)
            }
        }
    }

    function startDragging(dragHandleId: string, event: ResizeEvent): void {
        const handleElement = getResizeHandleElement(dragHandleId, groupElement)

        assert(handleElement, `Drag handle element not found for id "${dragHandleId}"`)

        const initialCursorPosition = getResizeEventCursorPosition(direction as PanelGroupDirection, event)

        $dragStateStore = {
            dragHandleId,
            dragHandleRect: handleElement.getBoundingClientRect(),
            initialCursorPosition,
            initialLayout: $layoutStore,
        }
    }

    function stopDragging() {
        $dragStateStore = null
    }

    function getPanelStyles(panelId: PanelId): Readable<string> {
        return derived([layoutStore, panelsStore, dragStateStore], ([layout, panels, dragState]) => {
            return computePanelFlexBoxStyle({
                dragState,
                layout,
                panels,
                panelIndex: findPanelDataIndex(panels, panelId),
            })
        })
    }

    function getPanelCollapsedState(panel: PanelInfo): Readable<boolean> {
        return derived([layoutStore, panelsStore], ([layout, panels]) => {
            const { panelSize, collapsible, collapsedSize = 0 } = getPanelMetadata(panels, panel, layout)

            return panelSize != null && collapsible === true && fuzzyNumbersEqual(panelSize, collapsedSize)
        })
    }

    setContext<PanelGroupContext>('panel-group-context', {
        groupId,
        registerPanel,
        getResizeHandler,
        startDragging,
        stopDragging,
        getPanelStyles,
        panelsStore,
        dragStateStore,
        getPanelGroupElement: () => groupElement,
        direction: direction as PanelGroupDirection,
        expandPanel,
        collapsePanel,
        getPanelCollapsedState,
    })
</script>

<div
    class="root"
    bind:this={groupElement}
    style:flex-direction={direction === PanelGroupDirection.Horizontal ? 'row' : 'column'}
    data-panel-group
    data-panel-group-id={groupId}
    data-panel-group-direction={direction}
>
    <slot />
</div>

<style lang="scss">
    .root {
        display: flex;
        width: 100%;
        height: 100%;
        overflow: hidden;
    }
</style>
