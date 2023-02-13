// Replication of style overwrites in CodeMirrorQueryInput.module.scss as
// CodeMirror theme

import { EditorView } from '@codemirror/view'

export const defaultTheme = EditorView.theme({
    '.cm-focus': {
        // Codemirror shows a focus ring by default. Since we handle that
        // differently, disable it here.
        outline: 'none',
    },
    '.cm-scroller': {
        // Codemirror shows a vertical scroll bar by default (when
        // overflowing). This disables it.
        overflowX: 'hidden',
    },
    '.cm-content': {
        caretColor: 'var(--search-query-text-color)',
        fontFamily: 'var(--code-font-family)',
        fontSize: 'var(--code-font-size)',
        color: 'var(--search-query-text-color)',
        // Disable default padding
        padding: 0,
    },
    '.cm-content.focus-visible,.cm-content:focus-visible': {
        boxShadow: 'none',
    },
    // @media (--xs-breakpoint-down) probably doesn't work here
    '.cm-line': {
        padding: 0,
    },
    '.cm-placeholder': {
        // CodeMirror uses display: inline-block by default, but that causes
        // Chrome to render a larger cursor if the placeholder holder spans
        // multiple lines. Firefox doesn't have this problem (but
        // setting display: inline doesn't seem to have a negative effect
        // either)
        display: 'inline',
        // Once again, Chrome renders the placeholder differently than
        // Firefox. CodeMirror sets 'word-break: break-word' (which is
        // deprecated) and 'overflow-wrap: anywhere' but they don't seem to
        // have an effect in Chrome (at least not in this instance).
        // Setting 'word-break: break-all' explicitly makes appearances a
        // bit better for example queries with long tokens.
        wordBreak: 'break-all',
    },
    '.cm-tooltip': {
        padding: '0.25rem',
        color: 'var(--search-query-text-color)',
        backgroundColor: 'var(--color-bg-1)',
        border: '1px solid var(--border-color)',
        borderRadius: 'var(--border-radius)',
        boxShadow: 'var(--box-shadow)',
        maxWidth: '50vw',
    },
    '.cm-tooltip p:last-child': {
        marginBottom: 0,
    },
    '.cm-toolip code': {
        backgroundColor: 'rgba(220, 220, 220, 0.4)',
        borderRadius: 'var(--border-radius)',
        padding: '0 0.4em',
    },
    '.cm-tooltip-section': {
        paddingBottom: '0.25rem',
        borderTopColor: 'var(--border-color)',
    },

    '.cm-tooltip-section:last-child': {
        paddingTop: '0.25rem',
        paddingBottom: 0,
    },
    '.cm-tooltip-section:last-child:first-child': {
        padding: 0,
    },
    '.cm-tooltip.cm-tooltip-autocomplete': {
        /* Resets padding added above to .cm-tooltip */
        padding: 0,
        color: 'var(--search-query-text-color)',
        backgroundColor: 'var(--color-bg-1)',
        // Default is 50vw
        maxWidth: '70vw',
        marginTop: '0.25rem', // Position is controlled absolutely but needs to be shifted down a bit from the default
    },
    // Requires some additional classes to overwrite default settings
    '.cm-tooltip.cm-tooltip-autocomplete > ul': {
        fontSize: 'var(--code-font-size)',
        fontFamily: 'var(--code-font-family)',
        maxHeight: '15rem',
    },
    '.cm-tooltip-autocomplete > ul > li': {
        alignItems: 'center',
        boxSizing: 'content-box',
        padding: '0.25rem 0.375rem',
        display: 'flex',
        height: '1.25rem',
    },
    '.cm-tooltip-autocomplete > ul > li[aria-selected]': {
        color: 'var(--search-query-text-color)',
        backgroundColor: 'var(--color-bg-2)',
    },
    '.cm-tooltip-autocomplete > ul > li[aria-selected] .tabStyle': {
        display: 'inline-block',
    },
    '.theme-dark .cm-tooltip-autocomplete > ul > li[aria-selected]': {
        backgroundColor: 'var(--color-bg-3)',
    },
    '.cm-tooltip-autocomplete svg': {
        flexShrink: 0,
    },
    '.cm-tooltip-autocomplete svg path': {
        fill: 'var(--search-query-text-color)',
    },
    '.cm-completionLabel': {
        flexShrink: 0,
    },
    '.cm-completionDetail': {
        paddingLeft: '0.25rem',
        fontSize: '0.675rem',
        color: 'var(--gray-06)',
        flex: 1,
        minWidth: 0,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        fontStyle: 'initial',
        textAlign: 'right',
    },
    '.cm-completionMatchedText': {
        // Reset
        textDecoration: 'none',

        // Our style
        color: 'var(--search-filter-keyword-color)',
        fontWeight: 'bold',
    },
})
