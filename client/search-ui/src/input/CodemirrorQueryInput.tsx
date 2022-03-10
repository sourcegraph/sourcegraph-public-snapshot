import { RangeSetBuilder } from '@codemirror/rangeset'
import { EditorState, EditorStateConfig, Extension, Facet, StateEffect, StateField } from '@codemirror/state'
import { hoverTooltip, TooltipView } from '@codemirror/tooltip'
import { EditorView, ViewUpdate, keymap, Decoration } from '@codemirror/view'
import React, { useEffect, useMemo, useRef, useState } from 'react'

import { renderMarkdown } from '@sourcegraph/common'
import { QueryChangeSource, SearchPatternType } from '@sourcegraph/search'
import { decorate, DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { toHover } from '@sourcegraph/shared/src/search/query/hover'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'

import styles from './CodemirrorQueryInput.module.scss'
import { MonacoQueryInputProps } from './MonacoQueryInput'

export const CodemirrorQueryInput: React.FunctionComponent<MonacoQueryInputProps> = ({
    patternType,
    selectedSearchContextSpec,
    queryState,
    onChange,
    onSubmit,
}) => {
    const container = useRef<HTMLDivElement | null>(null)

    const extensions = useMemo(
        () => [
            singleLine,
            submitOnEnter(onSubmit),
            notifyOnChange((value: string) =>
                onChange({
                    query: value,
                    changeSource: QueryChangeSource.userInput,
                })
            ),
            patternTypeField,
            parsedQueryFieldExtension,
            tokenHighlight,
            highlightFocusedFilter,
            tokenInfo,
        ],
        [onChange, onSubmit]
    )

    const view = useCodeMirror(container.current, queryState.query, extensions)

    // Update pattern type when it changes
    useEffect(() => {
        view?.dispatch({ effects: [patternTypeFieldEffect.of(patternType)] })
    }, [view, patternType])

    // Always focus the editor on selectedSearchContextSpec change
    useEffect(() => {
        if (selectedSearchContextSpec) {
            view?.focus()
        }
    }, [view, selectedSearchContextSpec])

    return <div ref={container} className={styles.root} />
}

function useCodeMirror(
    container: HTMLDivElement | null,
    value: string,
    extensions: EditorStateConfig['extensions'] = []
): EditorView | undefined {
    const [view, setView] = useState<EditorView>()

    useEffect(() => {
        if (container) {
            const view = new EditorView({
                state: EditorState.create({ doc: value ?? '', extensions }),
                parent: container,
            })
            setView(view)
            return () => {
                setView(undefined)
                view.destroy()
            }
        }
        return
        // Extensions and value are updated via transactions below
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [container])

    // Update editor value if necessary
    useEffect(() => {
        const currentValue = view?.state.doc.toString() ?? ''
        if (view && currentValue !== value) {
            view.dispatch({
                changes: { from: 0, to: currentValue.length, insert: value ?? '' },
            })
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [value])

    useEffect(() => {
        if (view) {
            view.dispatch({ effects: StateEffect.reconfigure.of(extensions) })
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [extensions])

    return view
}

// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
// The remainder of the file defines all the extensions that provide the query
// editor behavior.

// Enforces that the input won't split over multiple lines (basically prevents
// Enter from inserting a new line)
const singleLine = EditorState.transactionFilter.of(transaction => (transaction.newDoc.lines > 1 ? [] : transaction))

const submitOnEnter = (onSubmit: () => void): Extension =>
    keymap.of([
        {
            key: 'Enter',
            run: () => {
                onSubmit()
                return true
            },
        },
    ])

const notifyOnChange = (onChange: (value: string) => void): Extension =>
    EditorView.updateListener.of((update: ViewUpdate) => {
        if (update.docChanged) {
            // Looks like Text overwrites toString somehow
            // eslint-disable-next-line @typescript-eslint/no-base-to-string
            onChange(update.state.doc.toString())
        }
    })

// Defines decorators for syntax highlighting
type StyleNames = keyof typeof styles
const tokenDecorators: { [key: string]: Decoration } = Object.fromEntries(
    (Object.keys(styles) as StyleNames[]).map(style => [style, Decoration.mark({ class: styles[style] })])
)
const emptyDecorator = Decoration.mark({})
const focusedFilterDeco = Decoration.mark({ class: styles.focusedFilter })

// Chooses the correct decorator for the decorated token
const decoratedToDecoration = (token: DecoratedToken): Decoration => {
    let cssClass = 'identifier'
    switch (token.type) {
        case 'field':
        case 'whitespace':
        case 'keyword':
        case 'comment':
        case 'openingParen':
        case 'closingParen':
        case 'metaFilterSeparator':
        case 'metaRepoRevisionSeparator':
        case 'metaContextPrefix':
            cssClass = token.type
            break
        case 'metaPath':
        case 'metaRevision':
        case 'metaRegexp':
        case 'metaStructural':
        case 'metaPredicate':
            // The scopes value is derived from the token type and its kind.
            // E.g., regexpMetaDelimited derives from {@link RegexpMeta} and {@link RegexpMetaKind}.
            cssClass = `${token.type}${token.kind}`
            break
    }
    return tokenDecorators[cssClass] ?? emptyDecorator
}

// Editor state to keep information about the selected pattern type
const patternTypeField = StateField.define<SearchPatternType>({
    create() {
        return SearchPatternType.literal
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(patternTypeFieldEffect)) {
                return effect.value
            }
        }
        return value
    },
})
// Effect to update the the selected pattern type
const patternTypeFieldEffect = StateEffect.define<SearchPatternType>()

// Facet which parses the query using our existing parser. It depends on the current input
// (obviously) and the selected pattern type. It gets recomputed whenever one of
// those values changes.
// The parsed query is used for syntax highlighting and hover information.
const parsedQueryField = Facet.define<Token[], Token[]>({
    combine(input) {
        return input[0] ?? []
    },
})
const parsedQueryFieldExtension = parsedQueryField.compute(['doc', patternTypeField], state => {
    // Looks like Text overwrites toString somehow
    // eslint-disable-next-line @typescript-eslint/no-base-to-string
    const result = scanSearchQuery(state.doc.toString(), false, state.field(patternTypeField))
    return result.type === 'success' ? result.term : []
})

// This provides syntax highlighting. This is a custom solution so that we an
// use our existing query parser (instead of using codemirrors language
// support).
const tokenHighlight = EditorView.decorations.compute([parsedQueryField], state => {
    const query = state.facet(parsedQueryField)
    const builder = new RangeSetBuilder<Decoration>()
    for (const token of query) {
        for (const decoratedToken of decorate(token)) {
            builder.add(
                decoratedToken.range.start,
                decoratedToken.range.end + (decoratedToken.type === 'field' ? 1 : 0),
                decoratedToDecoration(decoratedToken)
            )
        }
    }
    return builder.finish()
})

// Tooltip information. This doesn't highlight the current token (yet).
const tokenInfo = hoverTooltip(
    (view, position) => {
        const tokensAtCursor = view.state
            .facet(parsedQueryField)
            ?.flatMap(decorate)
            .filter(({ range }) => range.start <= position && range.end > position)
        if (tokensAtCursor?.length === 0) {
            return null
        }
        const values: string[] = []
        let range: { start: number; end: number } | undefined
        tokensAtCursor.map(token => {
            switch (token.type) {
                case 'field': {
                    const resolvedFilter = resolveFilter(token.value)
                    if (resolvedFilter) {
                        values.push(
                            'negated' in resolvedFilter
                                ? resolvedFilter.definition.description(resolvedFilter.negated)
                                : resolvedFilter.definition.description
                        )
                        // Add 1 to end of range to include the ':'.
                        range = { start: token.range.start, end: token.range.end + 1 }
                    }
                    break
                }
                case 'pattern':
                case 'metaRevision':
                case 'metaRepoRevisionSeparator':
                case 'metaSelector':
                    values.push(toHover(token))
                    range = token.range
                    break
                case 'metaRegexp':
                case 'metaStructural':
                case 'metaPredicate':
                    values.push(toHover(token))
                    range = token.groupRange ? token.groupRange : token.range
                    break
            }
        })
        if (range) {
            return {
                pos: range.start,
                end: range.end,
                create(): TooltipView {
                    const dom = document.createElement('div')
                    dom.innerHTML = renderMarkdown(values.join(''))
                    return { dom }
                },
            }
        }
        return null
    },
    { hoverTime: 100 }
)

// Determines whether the cursor is over a filter and if yes, decorates that
// filter.
const highlightFocusedFilter = EditorView.decorations.compute(['selection', parsedQueryField], state => {
    const query = state.facet(parsedQueryField)
    const position = state.selection.main.head
    const focusedFilter = query.find(
        (token): token is Filter =>
            token.type === 'filter' && token.range.start <= position && token.range.end >= position
    )
    return focusedFilter
        ? Decoration.set(focusedFilterDeco.range(focusedFilter.range.start, focusedFilter.range.end))
        : Decoration.none
})
