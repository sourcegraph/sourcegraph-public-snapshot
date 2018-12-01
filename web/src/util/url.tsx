/**
 * Correctly handle use of meta/ctrl/alt keys during onClick events that open new pages
 */
export function openFromJS(path: string, event?: MouseEvent): void {
    if (event && (event.metaKey || event.altKey || event.ctrlKey || event.button === 1)) {
        window.open(path, '_blank')
    } else {
        window.location.href = path
    }
}
