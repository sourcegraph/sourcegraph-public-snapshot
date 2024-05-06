import type { PanelInfo, PanelsLayout } from '../types'

interface SerializedPanelGroupState {
    [panelIds: string]: PanelConfigurationState
}

export interface PanelConfigurationState {
    layout: number[]
    expandToSizes: {
        [panelId: string]: number
    }
}

/**
 * Important abstraction over key, it might be useful when we change logic
 * around persistent data to switch keys and hence don't think about migration
 * values from local storage which have been saved by old logic
 *
 * Changing key serialization resets existing saved panels information
 */
function getPanelIdKey(panelId: string): string {
    return `${panelId}-v2`
}

export function savePanelGroupLayout(
    panelId: string,
    panels: PanelInfo[],
    layout: PanelsLayout,
    panelSizesBeforeCollapse: Map<string, number>
): void {
    const panelsKey = getPanelsKey(panels)
    const state = loadSerializedPanelGroupState(panelId) ?? {}

    state[panelsKey] = {
        layout,
        expandToSizes: Object.fromEntries(panelSizesBeforeCollapse.entries()),
    }

    try {
        localStorage.setItem(getPanelIdKey(panelId), JSON.stringify(state))
    } catch (error) {
        console.error(error)
    }
}

export function loadPanelGroupLayout(groupId: string, panels: PanelInfo[]): PanelConfigurationState | null {
    const state = loadSerializedPanelGroupState(groupId) ?? {}
    const panelKey = getPanelsKey(panels)

    return state[panelKey] ?? null
}

function loadSerializedPanelGroupState(panelGroupKey: string): SerializedPanelGroupState | null {
    try {
        const serialized = localStorage.getItem(getPanelIdKey(panelGroupKey))
        if (serialized) {
            const parsed = JSON.parse(serialized)
            if (typeof parsed === 'object' && parsed !== undefined) {
                return parsed as SerializedPanelGroupState
            }
        }
    } catch {
        // Noop
    }

    return null
}

// Note that Panel ids might be user-provided (stable) or useId generated (non-deterministic)
// so they should not be used as part of the serialization key.
// Using the min/max size attributes should work well enough as a backup.
// Pre-sorting by minSize allows remembering layouts even if panels are re-ordered/dragged.
function getPanelsKey(panels: PanelInfo[]): string {
    return panels
        .map(panel => {
            const { constraints, idFromProps, order } = panel
            if (idFromProps) {
                return idFromProps
            }

            return order ? `${order}:${JSON.stringify(constraints)}` : JSON.stringify(constraints)
        })
        .sort((a, b) => a.localeCompare(b))
        .join(',')
}
