import { type Diagnostic as CMDiagnostic, linter, type LintSource } from '@codemirror/lint'
import { Compartment, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { renderMarkdown } from '@sourcegraph/common'
import { type Diagnostic, getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'

import { queryTokens } from './parsedQuery'

const theme = EditorView.theme({
    '.cm-diagnosticText': {
        display: 'block',
    },
    '.cm-diagnosticAction': {
        color: 'var(--body-color)',
        borderColor: 'var(--secondary)',
        backgroundColor: 'var(--secondary)',
        borderRadius: 'var(--border-radius)',
        padding: 'var(--btn-padding-y-sm) .5rem',
        fontSize: 'calc(min(0.75rem, 0.9166666667em))',
        lineHeight: '1rem',
        margin: '0.5rem 0 0 0',

        '& + .cm-diagnosticAction': {
            marginLeft: '1rem',
        },
    },
})

/**
 * Sets up client side query validation.
 */
export function queryDiagnostic(): Extension {
    // The setup is a bit "strange" because @codemirror/lint only triggers
    // linting when the document changes. But in our case the linting rules
    // change depending on the query "type" (regexp, structural, ...). Changing
    // the query type does not involve changing the document and to linting
    // wouldn't be triggered. To work around this we explictly reconfigure the
    // linter via a compartment when the parsed query changes but the document
    // hadsn't change. This queues a new linting pass.
    // See
    // - https://discuss.codemirror.net/t/can-we-manually-force-linting-even-if-the-document-hasnt-changed/3570/2
    // - https://github.com/sourcegraph/sourcegraph/issues/43836
    //
    const source: LintSource = view => {
        const query = view.state.facet(queryTokens)
        return query.tokens.length > 0 ? getDiagnostics(query.tokens, query.patternType).map(toCMDiagnostic) : []
    }
    const config = {
        delay: 200,
    }

    const linterCompartment = new Compartment()

    return [
        linterCompartment.of(linter(source, config)),
        EditorView.updateListener.of(update => {
            if (update.state.facet(queryTokens) !== update.startState.facet(queryTokens) && !update.docChanged) {
                update.view.dispatch({ effects: linterCompartment.reconfigure(linter(source, config)) })
            }
        }),
        theme,
    ]
}

function renderMarkdownNode(message: string): Element {
    const div = document.createElement('div')
    div.innerHTML = renderMarkdown(message)
    return div.firstElementChild || div
}

function toCMDiagnostic(diagnostic: Diagnostic): CMDiagnostic {
    return {
        from: diagnostic.range.start,
        to: diagnostic.range.end,
        message: diagnostic.message,
        renderMessage() {
            return renderMarkdownNode(diagnostic.message)
        },
        severity: diagnostic.severity,
        actions: diagnostic.actions?.map(action => ({
            name: action.label,
            apply(view) {
                view.dispatch({ changes: action.change, selection: action.selection })
                if (action.selection && !view.hasFocus) {
                    view.focus()
                }
            },
        })),
    }
}
