import { autocompletion, CompletionResult, Completion, snippet, completionKeymap } from '@codemirror/autocomplete'
import { RangeSetBuilder } from '@codemirror/rangeset'
import {
    EditorSelection,
    EditorState,
    EditorStateConfig,
    Extension,
    Facet,
    StateEffect,
    StateField,
    Prec,
} from '@codemirror/state'
import { hoverTooltip, TooltipView } from '@codemirror/tooltip'
import { EditorView, ViewUpdate, keymap, Decoration, placeholder as placeholderExtension } from '@codemirror/view'
import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import { editor as Monaco, MarkerSeverity, languages } from 'monaco-editor'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Observable, of } from 'rxjs'
import { delay, map, switchMap } from 'rxjs/operators'

import { renderMarkdown } from '@sourcegraph/common'
import { QueryChangeSource, SearchPatternType, SearchPatternTypeProps } from '@sourcegraph/search'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { getCompletionItems } from '@sourcegraph/shared/src/search/query/completion'
import { decorate, DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'
import { resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { toHover } from '@sourcegraph/shared/src/search/query/hover'
import { getSuggestionQuery } from '@sourcegraph/shared/src/search/query/providers'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './CodemirrorQueryInput.module.scss'
import { MonacoQueryInputProps } from './MonacoQueryInput'

const replacePattern = /[\n\r↵]/g

/**
 * This component provides a drop-in replacement for MonacoQueryInput. It
 * creates the approprate extensions and event handlers for the provided props.
 *
 * Deliberate differences compared to MonacoQueryInput:
 * - Filters are "highlighted" when the cursor is at their position
 * - Shift+Enter won't insert a new line if preventNewLine is true (default)
 * - Not supplying onSubmit and setting preventNewLine to false will result in a
 * new line being added when Enter is pressed
 */
export const CodemirrorMonacoFacade: React.FunctionComponent<MonacoQueryInputProps> = ({
    patternType,
    selectedSearchContextSpec,
    queryState,
    onChange,
    /**
     * If not provided and preventNewLine is false, Enter will insert a new
     * line. This is different from MonacoQueryInput's behavior.
     */
    onSubmit,
    autoFocus,
    onBlur,
    isSourcegraphDotCom,
    globbing,
    onHandleFuzzyFinder,
    onEditorCreated,
    interpretComments,
    isLightTheme,
    className,
    preventNewLine = true,
    placeholder,
}) => {
    const value = preventNewLine ? queryState.query.replace(replacePattern, '') : queryState.query
    const [editor, setEditor] = useState<EditorView | undefined>()
    const editorReference = useRef<EditorView>()

    const editorCreated = useCallback(
        editor => {
            setEditor(editor)
            editorReference.current = editor
            onEditorCreated?.(editor)
        },
        [editorReference, onEditorCreated]
    )

    const autocompletion = useMemo(
        () =>
            autocomplete(query => fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpec)), {
                globbing,
                isSourcegraphDotCom,
            }),
        [selectedSearchContextSpec, globbing, isSourcegraphDotCom]
    )

    const extensions = useMemo(() => {
        const extensions: Extension[] = [
            EditorView.updateListener.of((update: ViewUpdate) => {
                if (update.docChanged) {
                    onChange({
                        // Looks like Text overwrites toString somehow
                        // eslint-disable-next-line @typescript-eslint/no-base-to-string
                        query: update.state.doc.toString(),
                        changeSource: QueryChangeSource.userInput,
                    })
                }
                if (onBlur && update.focusChanged && !update.view.hasFocus) {
                    onBlur()
                }
            }),
            autocompletion,
        ]

        if (onSubmit) {
            extensions.push(Prec.highest(notifyOnEnter(onSubmit)))
        }

        if (onHandleFuzzyFinder) {
            extensions.push(keymap.of([{ key: 'Mod-k', run: () => (onHandleFuzzyFinder(true), true) }]))
        }

        if (preventNewLine) {
            extensions.push(singleLine)
        } else {
            // Automatically enable linewrapping in multi-line mode
            extensions.push(EditorView.lineWrapping)
        }

        if (placeholder) {
            extensions.push(placeholderExtension(placeholder))
        }
        return extensions
    }, [
        autocompletion,
        globbing,
        isLightTheme,
        isSourcegraphDotCom,
        onBlur,
        onChange,
        onHandleFuzzyFinder,
        onSubmit,
        placeholder,
        preventNewLine,
    ])

    // Always focus the editor on selectedSearchContextSpec change
    useEffect(() => {
        if (selectedSearchContextSpec) {
            editorReference.current?.focus()
        }
    }, [selectedSearchContextSpec])

    // Focus the editor if the autoFocus prop is truthy
    useEffect(() => {
        if (editor && autoFocus) {
            editor.focus()
        }
    }, [editor, autoFocus])

    useEffect(() => {
        if (!editor) {
            return
        }

        switch (queryState.changeSource) {
            case QueryChangeSource.userInput:
                // Don't react to user input
                break
            case QueryChangeSource.searchTypes:
            case QueryChangeSource.searchReference: {
                const selectionRange = queryState.selectionRange
                editor.dispatch({
                    selection: EditorSelection.range(selectionRange.start, selectionRange.end),
                    scrollIntoView: true,
                })
                /*
                if (queryState.showSuggestions) {
                    editor.trigger('triggerSuggestions', 'editor.action.triggerSuggest', {})
                }
                 */
                editor.focus()
                break
            }
            default: {
                // Place the cursor at the end of the query.
                editor.dispatch({
                    selection: EditorSelection.cursor(editor.state.doc.length),
                    scrollIntoView: true,
                })
            }
        }
    }, [editor, queryState])

    // It looks like <Shortcut ... /> needs a stable onMatch callback, hence we
    // are storing the editor in a ref so that `globalFocus` is stable.
    const globalFocus = useCallback(() => {
        if (
            editorReference.current &&
            !!document.activeElement &&
            !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)
        ) {
            editorReference.current.focus()
        }
    }, [editorReference])

    return (
        <>
            <CodemirrorQueryInput
                isLightTheme={isLightTheme}
                onEditorCreated={editorCreated}
                patternType={patternType}
                interpretComments={interpretComments}
                value={value}
                className={className}
                extensions={extensions}
            />
            {KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={globalFocus} />
            ))}
        </>
    )
}

interface CodeMirrorQueryInputProps extends ThemeProps, SearchPatternTypeProps {
    value: string
    onEditorCreated?: (editor: EditorView) => void
    // Whether comments are parsed and highlighted
    interpretComments?: boolean
    className?: string
    extensions: Extension[]
}

/**
 * "Core" codemirror query input component. Provides the basic behavior such as
 * theming, syntax highlighting and token info.
 */
const CodemirrorQueryInput: React.FunctionComponent<CodeMirrorQueryInputProps> = React.memo(
    ({ isLightTheme, onEditorCreated, patternType, interpretComments, value, className, extensions = [] }) => {
        const [container, setContainer] = useState<HTMLDivElement | null>(null)

        const editor = useCodeMirror(
            container,
            value,
            useMemo(
                () => [
                    EditorView.darkTheme.of(isLightTheme === false),
                    queryParsingOptions,
                    parsedQueryFieldExtension,
                    tokenHighlight,
                    queryDiagnostic,
                    tokenInfo,
                    highlightFocusedFilter,
                    ...extensions,
                ],
                [isLightTheme, extensions]
            )
        )

        useEffect(() => {
            if (editor) {
                onEditorCreated?.(editor)
            }
        }, [editor, onEditorCreated])

        // Update pattern type and/or interpretComments when the change
        useEffect(() => {
            editor?.dispatch({ effects: [setQueryOptions.of({ patternType, interpretComments })] })
        }, [editor, patternType, interpretComments])

        return <div ref={setContainer} className={classNames(styles.root, className)} id="monaco-query-input" />
    }
)

/**
 * Hook for rendering and updating a Codemirror instance.
 */
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

// Binds a function to the Enter key
const notifyOnEnter = (notify: () => void): Extension =>
    keymap.of([
        {
            key: 'Enter',
            run: () => {
                notify()
                return true
            },
        },
    ])

// Defines decorators for syntax highlighting
type StyleNames = keyof typeof styles
const tokenDecorators: { [key: string]: Decoration } = Object.fromEntries(
    (Object.keys(styles) as StyleNames[]).map(style => [style, Decoration.mark({ class: styles[style] })])
)
const emptyDecorator = Decoration.mark({})
const focusedFilterDeco = Decoration.mark({ class: styles.focusedFilter })

// Chooses the correct decorator for the decorated token. Copied (and adapated)
// from decoratedToMonaco (decoratedToken.ts).
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

// Editor state to keep information about how to parse the query
const queryParsingOptions = StateField.define<{ patternType: SearchPatternType; interpretComments?: boolean }>({
    create() {
        return {
            patternType: SearchPatternType.literal,
        }
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setQueryOptions)) {
                return { ...value, ...effect.value }
            }
        }
        return value
    },
})
// Effect to update the the selected pattern type
const setQueryOptions = StateEffect.define<{ patternType: SearchPatternType; interpretComments?: boolean }>()

interface ParsedQuery {
    patternType: SearchPatternType
    tokens: Token[]
}
// Facet which parses the query using our existing parser. It depends on the
// current input (obviously) and the selected pattern type. It gets recomputed
// whenever one of those values changes.
// The parsed query is used for syntax highlighting and hover information.
const parsedQueryField = Facet.define<ParsedQuery, ParsedQuery>({
    combine(input) {
        // There will always only be one extension which parses this query
        return input[0] ?? { patternType: SearchPatternType.literal, tokens: [] }
    },
})
const parsedQueryFieldExtension = parsedQueryField.compute(['doc', queryParsingOptions], state => {
    const { patternType, interpretComments } = state.field(queryParsingOptions)
    // Looks like Text overwrites toString somehow
    // eslint-disable-next-line @typescript-eslint/no-base-to-string
    const result = scanSearchQuery(state.doc.toString(), interpretComments, patternType)
    return {
        patternType,
        tokens: result.type === 'success' ? result.term : [],
    }
})

// This provides syntax highlighting. This is a custom solution so that we an
// use our existing query parser (instead of using codemirrors language
// support). That's not to say that we couldn't properly intergate with
// codemirror's language system with more effort.
const tokenHighlight = EditorView.decorations.compute([parsedQueryField], state => {
    const query = state.facet(parsedQueryField)
    const builder = new RangeSetBuilder<Decoration>()
    for (const token of query.tokens) {
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

// Determines whether the cursor is over a filter and if yes, decorates that
// filter.
const highlightFocusedFilter = EditorView.decorations.compute(['selection', parsedQueryField], state => {
    const query = state.facet(parsedQueryField)
    const position = state.selection.main.head
    const focusedFilter = query.tokens.find(
        (token): token is Filter =>
            token.type === 'filter' && token.range.start <= position && token.range.end >= position
    )
    return focusedFilter
        ? Decoration.set(focusedFilterDeco.range(focusedFilter.range.start, focusedFilter.range.end))
        : Decoration.none
})

// Tooltip information. This doesn't highlight the current token (yet).
const tokenInfo = hoverTooltip(
    (view, position) => {
        const tokensAtCursor = view.state
            .facet(parsedQueryField)
            .tokens?.flatMap(decorate)
            .filter(token => isTokenInRange(position, token))
        if (tokensAtCursor?.length === 0) {
            return null
        }
        const values: string[] = []
        let range: { start: number; end: number } | undefined

        // Copied and adapated from getHoverResult (hover.ts)
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
                        // Add 3 to end of range to include the ':'.
                        // (there seems to be a bug with computing the correct
                        // range end)
                        range = { start: token.range.start, end: token.range.end + 3 }
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

// Hooks query diagnostics into the editor.
// The facet stores the diagnostics data which is used by the text decoration
// extension and the tooltip extensions respectively.
const diagnostics = Facet.define<Monaco.IMarkerData[], Monaco.IMarkerData[]>({
    combine: markerData => markerData.flat(),
})
const diagnosticDecos: { [key in MarkerSeverity]: Decoration } = {
    [MarkerSeverity.Hint]: emptyDecorator,
    [MarkerSeverity.Info]: emptyDecorator,
    [MarkerSeverity.Warning]: Decoration.mark({ class: styles.diagnosticWarning }),
    [MarkerSeverity.Error]: Decoration.mark({ class: styles.diagnosticError }),
}
const queryDiagnostic: Extension[] = [
    // Compute diagnostics when query changes
    diagnostics.compute([parsedQueryField], state => {
        const query = state.facet(parsedQueryField)
        return query.tokens.length > 0 ? getDiagnostics(query.tokens, query.patternType) : []
    }),
    // Generate diagnostic markers
    EditorView.decorations.compute([diagnostics], state =>
        Decoration.set(
            state
                .facet(diagnostics)
                .map(marker => diagnosticDecos[marker.severity].range(marker.startColumn - 1, marker.endColumn - 1)),
            true
        )
    ),
    // Show diagnostic message on hover
    hoverTooltip(
        (view, position) => {
            const markersAtCursor = view.state
                .facet(diagnostics)
                .filter(({ startColumn, endColumn }) => startColumn - 1 <= position && endColumn > position)
            if (markersAtCursor?.length === 0) {
                return null
            }

            return {
                // TODO: Properly compute range for multiple markers
                pos: markersAtCursor[0].startColumn - 1,
                end: markersAtCursor[0].endColumn - 1,
                create(): TooltipView {
                    const dom = document.createElement('div')
                    dom.innerHTML = renderMarkdown(markersAtCursor.map(marker => marker.message).join('\n\n'))
                    return { dom }
                },
            }
        },
        { hoverTime: 100 }
    ),
]

// Hook up autocompletion
const autocomplete = (
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>,
    options: { globbing: boolean; isSourcegraphDotCom: boolean }
): Extension[] => [
    // Uses the default keymapping by changes accepting suggestions from Enter
    // to Tab
    Prec.highest(
        keymap.of(
            completionKeymap.map(keybinding =>
                keybinding.key === 'Enter' ? { ...keybinding, key: 'Tab' } : keybinding
            )
        )
    ),
    autocompletion({
        defaultKeymap: false,
        override: [
            context => {
                const query = context.state.facet(parsedQueryField)
                const token = query.tokens.find(token => isTokenInRange(context.pos - 1, token))
                if (!token) {
                    return null
                }
                return of(getSuggestionQuery(query.tokens, token))
                    .pipe(
                        // We use a delay here to implement a custom debounce. In the
                        // next step we check if the current completion request was
                        // cancelled in the meantime (`context.aborted`).
                        // This prevents us from needlessly running multiple suggestion
                        // queries.
                        delay(200),
                        switchMap(query =>
                            context.aborted
                                ? Promise.resolve(null)
                                : getCompletionItems(
                                      token,
                                      { column: context.pos + 1 },
                                      fetchSuggestions(query),
                                      options.globbing,
                                      options.isSourcegraphDotCom
                                  )
                        ),
                        map((completionList): CompletionResult | null => {
                            if (completionList === null || completionList.suggestions.length === 0) {
                                return null
                            }
                            return {
                                from:
                                    token.type === 'filter'
                                        ? token.value?.range.start ?? context.pos
                                        : token.range.start,
                                options: toCMCompletions(completionList),
                            }
                        })
                    )
                    .toPromise()
            },
        ],
    }),
]

function toCMCompletions(completionList: languages.CompletionList): Completion[] {
    // Boost suggestions by position because it appears they are already orderd.
    let boost = 99
    return completionList.suggestions.map(
        (item): Completion => ({
            type: languages.CompletionItemKind[item.kind].toLowerCase(),
            label: typeof item.label === 'string' ? item.label : item.label.name,
            apply:
                ((item.insertTextRules ?? 0) & languages.CompletionItemInsertTextRule.InsertAsSnippet) ===
                languages.CompletionItemInsertTextRule.InsertAsSnippet
                    ? snippet(item.insertText)
                    : item.insertText,
            detail: item.detail,
            info: item.documentation?.toString(),
            boost: boost--,
        })
    )
}

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function isTokenInRange(
    position: number,
    token: { type: DecoratedToken['type']; range: { start: number; end: number } }
): boolean {
    return token.range.start <= position && token.range.end + (token.type === 'field' ? 2 : 0) > position
}
