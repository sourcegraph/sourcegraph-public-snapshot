import { getResizeHandleElementsForGroup } from './getResizeHandleElementsForGroup'

export function getResizeHandleElementIndex(
    groupId: string,
    id: string,
    scope: ParentNode | HTMLElement = document
): number | null {
    const handles = getResizeHandleElementsForGroup(groupId, scope)
    const index = handles.findIndex(handle => handle.getAttribute('data-panel-resize-handle-id') === id)
    return index ?? null
}
