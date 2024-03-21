export function getResizeHandleElementsForGroup(
    groupId: string,
    scope: ParentNode | HTMLElement = document
): HTMLElement[] {
    return Array.from(scope.querySelectorAll(`[data-panel-resize-handle-id][data-panel-group-id="${groupId}"]`))
}
