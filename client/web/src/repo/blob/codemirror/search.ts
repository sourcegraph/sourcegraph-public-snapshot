/**
 * This extension exends CodeMirro's own search extension with a custom search
 * UI.
 */

import {
    findNext,
    findPrevious,
    getSearchQuery,
    search as codemirrorSearch,
    searchKeymap,
    SearchQuery,
    setSearchQuery,
} from '@codemirror/search'
import { Extension } from '@codemirror/state'
import { EditorView, keymap, Panel, runScopeHandlers, ViewUpdate } from '@codemirror/view'
import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'
import { getButtonClassName, getLabelClassName } from '@sourcegraph/wildcard'

import { createElement } from '../../../util/dom'

const buttonClassName = getButtonClassName({ size: 'sm', outline: true, variant: 'secondary' })
const labelClassName = getLabelClassName({ size: 'small', mode: 'single-line' })

class SearchPanel implements Panel {
    public dom: HTMLElement
    public top = true
    private query: SearchQuery
    private input: HTMLInputElement
    private caseSensitive: HTMLInputElement
    private regexp: HTMLInputElement

    constructor(private view: EditorView) {
        const previous = createElement(
            'button',
            {
                type: 'button',
                className: [buttonClassName, 'mr-2'].join(' '),
                onclick: () => findPrevious(view),
            },
            createSVGIcon(mdiChevronUp),
            'Previous'
        )
        previous.setAttribute('data-testid', 'blob-view-search-next')
        const next = createElement(
            'button',
            {
                type: 'button',
                className: buttonClassName,
                onclick: () => findNext(view),
            },
            createSVGIcon(mdiChevronDown),
            'Next'
        )
        next.setAttribute('data-testid', 'blob-view-search-previous')

        this.input = createElement('input', {
            name: 'search',
            placeholder: 'Find...',
            className: 'form-control form-control-sm mr-2',
            autocomplete: 'off',
            onchange: this.commit,
            onkeyup: this.commit,
        })
        this.input.setAttribute('main-field', 'true')

        this.caseSensitive = createElement('input', { type: 'checkbox', className: 'mr-2', onchange: this.commit })
        this.caseSensitive.setAttribute('data-testid', 'blob-view-search-case-sensitive')
        this.regexp = createElement('input', { type: 'checkbox', className: 'mr-2', onchange: this.commit })
        this.regexp.setAttribute('data-testid', 'blob-view-search-regexp')

        this.dom = createElement(
            'div',
            { className: 'search-container', onkeydown: this.onkeydown },
            this.input,
            previous,
            next,
            createElement(
                'label',
                { className: `form-check-label mx-2 ml-3 ${labelClassName}` },
                this.caseSensitive,
                'Match case'
            ),
            createElement('label', { className: `form-check-label mx-2 ${labelClassName}` }, this.regexp, 'Regexp')
        )

        this.query = getSearchQuery(this.view.state)

        this.setQuery(this.query)
    }

    public update(update: ViewUpdate): void {
        const currentQuery = getSearchQuery(update.state)
        if (!currentQuery.eq(getSearchQuery(update.startState))) {
            this.setQuery(currentQuery)
        }
    }

    public mount(): void {
        this.input.focus()
        this.input.select()
    }

    // Taken from CodeMirror's default serach panel implementation. This is
    // necessary so that pressing Meta+F (and other CodeMirror keybindings) will
    // trigger the configured event handlers and not just fall back to the
    // browser's default behavior.
    private onkeydown = (event: KeyboardEvent): void => {
        if (runScopeHandlers(this.view, event, 'search-panel')) {
            event.preventDefault()
        } else if (event.keyCode === 13 && event.target === this.input) {
            event.preventDefault()
            if (event.shiftKey) {
                findPrevious(this.view)
            } else {
                findNext(this.view)
            }
        }
    }

    private commit = (): void => {
        const query = new SearchQuery({
            search: this.input.value,
            caseSensitive: this.caseSensitive.checked,
            regexp: this.regexp.checked,
        })

        if (!query.eq(this.query)) {
            this.view.dispatch({ effects: setSearchQuery.of(query) })

            // The following code scrolls next match into view if there is no
            // match in the visible viewport. This is done by searching for the
            // text from the currently top visible line and determining whether
            // the next match is in the current viewport

            const { scrollTop } = this.view.scrollDOM

            // Get top visible line. More than half of the line must be visible.
            // We don't use `view.viewportLineBlocks` because that also includes
            // lines that are rendered but not actually visible.
            let topLineBlock = this.view.lineBlockAtHeight(scrollTop)
            if (Math.abs(topLineBlock.bottom - scrollTop) <= topLineBlock.height / 2) {
                topLineBlock = this.view.lineBlockAtHeight(scrollTop + topLineBlock.height)
            }

            let result = getSearchQuery(this.view.state).getCursor(this.view.state.doc, topLineBlock.from).next()
            if (result.done) {
                // No match in the remainder of the document, wrap around
                result = getSearchQuery(this.view.state).getCursor(this.view.state.doc).next()
                if (result.done) {
                    // Search term is not in the document, nothing to do
                    return
                }
            }

            const matchLineBlock = this.view.lineBlockAt(result.value.from)
            const matchLineCenter = matchLineBlock.top + matchLineBlock.height / 2

            if (matchLineCenter < scrollTop || matchLineCenter > scrollTop + this.view.scrollDOM.clientHeight) {
                this.view.dispatch({
                    effects: EditorView.scrollIntoView(result.value.from, {
                        y: 'center',
                    }),
                })
            }
        }
    }

    private setQuery(query: SearchQuery): void {
        this.query = query
        this.input.value = this.query.search
        this.caseSensitive.checked = this.query.caseSensitive
        this.regexp.checked = this.query.regexp
    }
}

export const search: Extension = [
    EditorView.theme({
        '.search-container': {
            backgroundColor: 'var(--code-bg)',
            display: 'flex',
            alignItems: 'center',
            padding: '0.375rem 1rem',
        },
        '.search-container > input.form-control': {
            width: '15rem',
        },
        '.search-container input[type="checkbox"]': {
            verticalAlign: 'text-bottom',
        },
        '.search-container svg': {
            width: 'var(--icon-inline-size)',
            height: 'var(--icon-inline-size)',
            boxSizing: 'border-box',
            textAlign: 'center',
            marginRight: '0.25rem',
            // The icons contain whitespace themselves, this makes the button
            // look more centered
            marginLeft: '-0.125rem',
            verticalAlign: 'text-bottom',
        },
        '.cm-searchMatch': {
            backgroundColor: 'var(--mark-bg)',
        },
        '.cm-searchMatch-selected': {
            backgroundColor: 'var(--oc-orange-3)',
        },
    }),
    keymap.of(searchKeymap),
    codemirrorSearch({
        createPanel: view => new SearchPanel(view),
    }),
]
