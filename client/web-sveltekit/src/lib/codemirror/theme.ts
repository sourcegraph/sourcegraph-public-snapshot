import { EditorView } from '@codemirror/view'

export const defaultTheme = EditorView.baseTheme({
    // Overwrites the default cursor color, which has too low contrast in dark mode
    '.theme-dark & .cm-cursor': {
        borderLeftColor: 'var(--grey-07)',
    },
})
