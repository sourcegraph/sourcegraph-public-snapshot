import { useMemo } from 'react'

import { EditorState, Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { useLocation } from 'react-router-dom'

import { CodeMirrorEditor, defaultEditorTheme } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'

import { selectableLineNumbers } from '../../repo/blob/codemirror/linenumbers'

export const IngestedFileViewer: React.FunctionComponent<{ contents: string }> = ({ contents }) => {
    const isLightTheme = useIsLightTheme()

    const location = useLocation()

    const lineNumber = useMemo(() => {
        const params = new URLSearchParams(location.search)
        const line = params.get('L')
        return line ? parseInt(line, 10) : undefined
    }, [location.search])

    const extensions: Extension[] = useMemo(
        () => [
            EditorView.darkTheme.of(isLightTheme === false),
            EditorState.readOnly.of(true),
            EditorView.theme({
                '.selected-line, .cm-line.selected-line': {
                    backgroundColor: 'var(--code-selection-bg)',
                },
                '.cm-lineNumbers .cm-gutterElement:hover': {
                    textDecoration: 'none',
                    cursor: 'auto',
                },
                '.cm-scroller': {
                    borderRadius: 'var(--border-radius)',
                },
            }),
            selectableLineNumbers({
                onSelection: () => {},
                initialSelection: lineNumber ? { line: lineNumber } : null,
                navigateToLineOnAnyClick: true,
            }),
            defaultEditorTheme,
        ],
        [isLightTheme, lineNumber]
    )

    return <CodeMirrorEditor value={contents} extensions={extensions} />
}
