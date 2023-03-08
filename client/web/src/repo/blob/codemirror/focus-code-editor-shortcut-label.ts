import { EditorView, showPanel, ViewUpdate } from '@codemirror/view'

function createShortcutHelpPanel(view: EditorView) {
    let dom = document.createElement('kbd')
    dom.textContent = 'C'
    dom.classList.add('shortcut-help-panel')

    return {
        top: true,
        dom,
        update(update: ViewUpdate) {
            if (update.heightChanged) {
                if (view.scrollDOM.clientHeight && view.scrollDOM.clientHeight < view.contentHeight) {
                    dom.classList.add('shortcut-help-panel--with-scrollbar')
                } else {
                    dom.classList.remove('shortcut-help-panel--with-scrollbar')
                }
            }
        },
    }
}

/**
 * Extension adding focus code editor shortcut label to the end of the first line.
 */
export const focusCodeEditorShortcutLabel = (enableBlobPageSwitchAreasShortcuts?: boolean) => {
    if (!enableBlobPageSwitchAreasShortcuts) {
        return []
    }

    return [
        showPanel.of(createShortcutHelpPanel),
        EditorView.theme({
            '.shortcut-help-panel': {
                position: 'absolute',
                right: '0.5rem',
                top: '0.5rem',
                display: 'inline-flex',
                alignItems: 'center',
                color: 'var(--text-muted)',
            },
            '.shortcut-help-panel--with-scrollbar': {
                right: '1.25rem',
            },
        }),
    ]
}
