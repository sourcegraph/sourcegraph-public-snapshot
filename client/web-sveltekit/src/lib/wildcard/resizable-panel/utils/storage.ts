import type { PanelInfo, PanelsLayout } from '../types'

interface SerializedPanelGroupState {
    [panelIds: string]: PanelsLayout
}

export function savePanelGroupLayout(panelId: string, panels: PanelInfo[], layout: PanelsLayout): void {
    const panelsKey = getPanelsKey(panels)
    const state = loadSerializedPanelGroupState(panelId) ?? {}

    state[panelsKey] = layout

    try {
        localStorage.setItem(panelId, JSON.stringify(state))
    } catch (error) {
        console.error(error)
    }
}

export function loadPanelGroupLayout(groupId: string, panels: PanelInfo[]): PanelsLayout | null {
    const state = loadSerializedPanelGroupState(groupId) ?? {}
    const panelKey = getPanelsKey(panels)

    return state[panelKey] ?? null
}

function loadSerializedPanelGroupState(panelGroupKey: string): SerializedPanelGroupState | null {
    try {
        const serialized = localStorage.getItem(panelGroupKey)
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
