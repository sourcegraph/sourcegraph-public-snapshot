import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { closeCompletion, startCompletion } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import {
    EditorSelection,
    Extension,
    Facet,
    StateEffect,
    StateField,
    Prec,
    RangeSetBuilder,
    MapMode,
    Compartment,
    Range,
} from '@codemirror/state'
import {
    EditorView,
    ViewUpdate,
    keymap,
    Decoration,
    placeholder as placeholderExtension,
    ViewPlugin,
    hoverTooltip,
    TooltipView,
    WidgetType,
} from '@codemirror/view'
import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import { editor as Monaco, MarkerSeverity } from 'monaco-editor'

import { renderMarkdown } from '@sourcegraph/common'
import { EditorHint, QueryChangeSource, SearchPatternTypeProps } from '@sourcegraph/search'
import { useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { DecoratedToken, toCSSClassName } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'
import { resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { toHover } from '@sourcegraph/shared/src/search/query/hover'
import { Node } from '@sourcegraph/shared/src/search/query/parser'
import { Filter, KeywordKind } from '@sourcegraph/shared/src/search/query/token'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { createDefaultSuggestions, createUpdateableField, singleLine } from './extensions'
import {
    decoratedTokens,
    queryTokens,
    parseInputAsQuery,
    setQueryParseOptions,
    parsedQuery,
} from './extensions/parsedQuery'
import { MonacoQueryInputProps } from './MonacoQueryInput'

import styles from './CodeMirrorQueryInput.module.scss'

/**
 * This component provides a drop-in replacement for MonacoQueryInput. It
 * creates the appropriate extensions and event handlers for the provided props.
 *
 * Deliberate differences compared to MonacoQueryInput:
 * - Filters are "highlighted" when the cursor is at their position
 * - Shift+Enter won't insert a new line if 'preventNewLine' is true (default)
 * - Not supplying 'onSubmit' and setting 'preventNewLine' to false will result
 * in a new line being added when Enter is pressed
 */
export const CodeMirrorMonacoFacade: React.FunctionComponent<React.PropsWithChildren<MonacoQueryInputProps>> = ({
    patternType,
    selectedSearchContextSpec,
    queryState,
    onChange,
    onSubmit,
    autoFocus,
    onFocus,
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
    editorOptions,
    ariaLabel = 'Search query',
    // Used by the VSCode extension (which doesn't use this component directly,
    // but added for future compatibility)
    fetchStreamSuggestions = defaultFetchStreamSuggestions,
    onCompletionItemSelected,
    // Not supported:
    // editorClassName: This only seems to be used by MonacoField to position
    // placeholder text properly. CodeMirror has built-in support for
    // placeholders.
}) => {
    // We use both, state and a ref, for the editor instance because we need to
    // re-run some hooks when the editor changes but we also need a stable
    // reference that doesn't change across renders (and some hooks should only
    // run when a prop changes, not the editor).
    const [editor, setEditor] = useState<EditorView | undefined>()
    const editorReference = useRef<EditorView>()

    const editorCreated = useCallback(
        (editor: EditorView) => {
            setEditor(editor)
            editorReference.current = editor
            onEditorCreated?.(editor)
        },
        [editorReference, onEditorCreated]
    )

    const autocompletion = useMemo(
        () =>
            createDefaultSuggestions({
                fetchSuggestions: query =>
                    fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpec)),
                globbing,
                isSourcegraphDotCom,
            }),
        [selectedSearchContextSpec, globbing, isSourcegraphDotCom, fetchStreamSuggestions]
    )

    const extensions = useMemo(() => {
        const extensions: Extension[] = [
            EditorView.contentAttributes.of({ 'aria-label': ariaLabel }),
            callbacksField,
            autocompletion,
        ]

        if (preventNewLine) {
            // NOTE: If a submit handler is assigned to the query input then the pressing
            // enter won't insert a line break anyway. In that case, this exnteions ensures
            // that line breaks are stripped from pasted input.
            extensions.push(singleLine)
        } else {
            // Automatically enable line wrapping in multi-line mode
            extensions.push(EditorView.lineWrapping)
        }

        if (placeholder) {
            extensions.push(placeholderExtension(placeholder))
        }

        if (editorOptions?.readOnly) {
            extensions.push(EditorView.editable.of(false))
        }
        return extensions
    }, [ariaLabel, autocompletion, placeholder, preventNewLine, editorOptions])

    // Update callback functions via effects. This avoids reconfiguring the
    // whole editor when a callback changes.
    useEffect(() => {
        if (editor) {
            setCallbacks(editor, {
                onChange,
                onSubmit,
                onFocus,
                onBlur,
                onCompletionItemSelected,
                onHandleFuzzyFinder,
            })
        }
    }, [editor, onChange, onSubmit, onFocus, onBlur, onCompletionItemSelected, onHandleFuzzyFinder])

    // Always focus the editor on 'selectedSearchContextSpec' change
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

    // Update the editor's selection and cursor depending on how the search
    // query was changed.
    useEffect(() => {
        if (!editor) {
            return
        }

        if (queryState.changeSource === QueryChangeSource.userInput) {
            // Don't react to user input
            return
        }

        editor.dispatch({
            selection: queryState.selectionRange
                ? // Select the specified range (most of the time this will be a
                  // placeholder filter value).
                  EditorSelection.range(queryState.selectionRange.start, queryState.selectionRange.end)
                : // Place the cursor at the end of the query.
                  EditorSelection.cursor(editor.state.doc.length),
            scrollIntoView: true,
        })

        if (queryState.hint) {
            if ((queryState.hint & EditorHint.Focus) === EditorHint.Focus) {
                editor.focus()
            }
            if ((queryState.hint & EditorHint.ShowSuggestions) === EditorHint.ShowSuggestions) {
                startCompletion(editor)
            }
        }
    }, [editor, queryState])

    // It looks like <Shortcut ... /> needs a stable onMatch callback, hence we
    // are storing the editor in a ref so that `globalFocus` is stable.
    const globalFocus = useCallback(() => {
        if (editorReference.current && !!document.activeElement && !isInputElement(document.activeElement)) {
            editorReference.current.focus()
        }
    }, [editorReference])

    return (
        <>
            <CodeMirrorQueryInput
                isLightTheme={isLightTheme}
                onEditorCreated={editorCreated}
                patternType={patternType}
                interpretComments={interpretComments}
                value={queryState.query}
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
export const CodeMirrorQueryInput: React.FunctionComponent<
    React.PropsWithChildren<CodeMirrorQueryInputProps>
> = React.memo(
    ({ isLightTheme, onEditorCreated, patternType, interpretComments, value, className, extensions = [] }) => {
        // This is using state instead of a ref because `useRef` doesn't cause a
        // re-render when the ref is attached, but we need that so that
        // `useCodeMirror` is called again and the editor is actually created.
        // See https://reactjs.org/docs/hooks-faq.html#how-can-i-measure-a-dom-node
        const [container, setContainer] = useState<HTMLDivElement | null>(null)
        const externalExtensions = useMemo(() => new Compartment(), [])
        const themeExtension = useMemo(() => new Compartment(), [])

        const editor = useCodeMirror(
            container,
            value,
            useMemo(
                () => [
                    keymap.of(historyKeymap),
                    keymap.of(defaultKeymap),
                    history(),
                    themeExtension.of(EditorView.darkTheme.of(isLightTheme === false)),
                    parseInputAsQuery({ patternType, interpretComments }),
                    querySyntaxHighlighting,
                    queryDiagnostic,
                    tokenInfo(),
                    highlightFocusedFilter,
                    externalExtensions.of(extensions),
                ],
                // patternType and interpretComments are updated via a
                // transaction since there is no need to re-initialize all
                // extensions
                // The extensions passed in via `extensions` are update via a
                // compartment
                // The theme (`isLightTheme`) is also updated via a compartment
                // eslint-disable-next-line react-hooks/exhaustive-deps
                [themeExtension, externalExtensions]
            )
        )

        // Notify parent component about editor instance. Among other things,
        // having a reference to the editor allows other components to initiate
        // transactions.
        useEffect(() => {
            if (editor) {
                onEditorCreated?.(editor)
            }
        }, [editor, onEditorCreated])

        // Update pattern type and/or interpretComments when changed
        useEffect(() => {
            editor?.dispatch({ effects: setQueryParseOptions.of({ patternType, interpretComments }) })
        }, [editor, patternType, interpretComments])

        // Update theme if it changes
        useEffect(() => {
            editor?.dispatch({ effects: themeExtension.reconfigure(EditorView.darkTheme.of(isLightTheme === false)) })
        }, [editor, themeExtension, isLightTheme])

        // Update external extensions if they changed
        useEffect(() => {
            editor?.dispatch({ effects: externalExtensions.reconfigure(extensions) })
        }, [editor, externalExtensions, extensions])

        return (
            <div
                ref={setContainer}
                className={classNames(styles.root, className)}
                data-test-id="codemirror-query-input"
            />
        )
    }
)

// The remainder of the file defines all the extensions that provide the query
// editor behavior. Here is also a brief overview over CodeMirror's architecture
// to make more sense of this (see https://codemirror.net/6/docs/guide/ for more
// details):
//
// In its own words, CodeMirror has a "Functional Core" and an "Imperative
// Shell". Updates to the editor's state are performed via transactions and
// produce a new state. This new state is passed to the editor's view, which in
// turn updates itself according to the new state.
//
// Almost all behavior is implemented via extensions. There are various types of
// extensions and sometimes multiple different types are needed to implement
// some functionality:
//
// - Transaction filters can inspect transactions to change them or remove them.
// - StateFields store arbitrary values, often some form of configuration. The
// value of a state field can be updated via transactions. The fields can be
// accessed anywhere where there is access to the field object and to the
// editor's state.
// - StateEffects are a way to update a StateField's value. A StateField can
// inspect a transaction to see whether it has a specific effect and extact the
// value from the effect.
// This implementation uses this to update the query parsing options that are
// passed to the editor from the outer component.
// - Facets provide external extension points and allow aggregation of multiple
// inputs into a single output value. They also make it possible to derive
// values from the editor's current state. A concrete example is the built-in
// `EditorView.decorations` facet. A new extension can be created by calling the
// facets `compute` method, which will compute a new list of decorations based
// on the editor's current state.
// This is how the query syntax highlighting and diagnostic decorations are
// implemented.
// As another example, the diagnostics information is also stored in a facet.
// We could eventually have multiple sources that compute diagnostics
// information (which gets combined into a single list). The extension which
// computes the diagnostics decorations doesn't (need to) know where all the
// information came from.
// - ViewPlugins provide a way to extend the editor's UI in various ways. They
// can access the editor's current state and update decorations or DOM elements
// based on it.
//
// Sometimes it's not always obvious which type of extension to use to achieve a
// certain goal (and I don't claim that the implementation below is optimal).

// Instead of deriving extensions directly form props, these event handlers are
// configured via a field. This means that their values can be updated via
// transactions instead of having to reconfigure the whole editor. This is
// especially useful if the event handlers are not stable across re-renders.
// Instead of creating a separate field for every handler, all handlers are set
// via a single field to keep complexity manageable.
const [callbacksField, setCallbacks] = createUpdateableField<
    Pick<
        MonacoQueryInputProps,
        'onChange' | 'onSubmit' | 'onFocus' | 'onBlur' | 'onCompletionItemSelected' | 'onHandleFuzzyFinder'
    >
>(
    callbacks => [
        Prec.high(
            keymap.of([
                {
                    key: 'Enter',
                    run: view => {
                        const { onSubmit } = view.state.field(callbacks)
                        if (onSubmit) {
                            // Cancel/close any open completion popovers
                            closeCompletion(view)
                            onSubmit()
                            return true
                        }
                        return false
                    },
                },
            ])
        ),
        keymap.of([
            {
                key: 'Mod-k',
                run: view => {
                    const { onHandleFuzzyFinder } = view.state.field(callbacks)
                    if (onHandleFuzzyFinder) {
                        onHandleFuzzyFinder(true)
                        return true
                    }
                    return false
                },
            },
        ]),
        EditorView.updateListener.of((update: ViewUpdate) => {
            const { state, view } = update
            const { onChange, onFocus, onBlur, onCompletionItemSelected } = state.field(callbacks)

            if (update.docChanged) {
                onChange({
                    query: state.sliceDoc(),
                    changeSource: QueryChangeSource.userInput,
                })
            }

            // The focus and blur event handlers are implemented via state update handlers
            // because it appears that binding them as DOM event handlers triggers them at
            // the moment they are bound if the editor is already in that state ((not)
            // focused). See https://github.com/sourcegraph/sourcegraph/issues/37721#issuecomment-1166300433
            if (update.focusChanged) {
                if (view.hasFocus) {
                    onFocus?.()
                } else {
                    closeCompletion(view)
                    onBlur?.()
                }
            }
            if (
                onCompletionItemSelected &&
                update.transactions.some(transaction => transaction.isUserEvent('input.complete'))
            ) {
                onCompletionItemSelected()
            }
        }),
    ],
    { onChange: () => {} }
)

// Defines decorators for syntax highlighting
const tokenDecorators: { [key: string]: Decoration } = {}
const emptyDecorator = Decoration.mark({})
const focusedFilterDeco = Decoration.mark({ class: styles.focusedFilter })

// Chooses the correct decorator for the decorated token
const decoratedToDecoration = (token: DecoratedToken): Decoration => {
    const className = toCSSClassName(token)
    const decorator = tokenDecorators[className]
    return decorator || (tokenDecorators[className] = Decoration.mark({ class: className }))
}

// This provides syntax highlighting. This is a custom solution so that we an
// use our existing query parser (instead of using CodeMirror's language
// support). That's not to say that we couldn't properly integrate with
// CodeMirror's language system with more effort.
const querySyntaxHighlighting = EditorView.decorations.compute([decoratedTokens], state => {
    const tokens = state.facet(decoratedTokens)
    const builder = new RangeSetBuilder<Decoration>()
    for (const token of tokens) {
        builder.add(token.range.start, getEndPosition(token), decoratedToDecoration(token))
    }
    return builder.finish()
})

class PlaceholderWidget extends WidgetType {
    constructor(private placeholder: string) {
        super()
    }

    /* eslint-disable-next-line id-length */
    public eq(other: PlaceholderWidget): boolean {
        return this.placeholder === other.placeholder
    }

    public toDOM(): HTMLElement {
        const span = document.createElement('span')
        span.className = styles.placeholder
        span.textContent = this.placeholder
        return span
    }
}

// Determines whether the cursor is over a filter and if yes, decorates that
// filter.
const highlightFocusedFilter = ViewPlugin.define(
    () => ({
        decorations: Decoration.none,
        update(update) {
            if (update.docChanged || update.selectionSet || update.focusChanged) {
                if (update.view.hasFocus) {
                    const query = update.state.facet(queryTokens)
                    const position = update.state.selection.main.head
                    const focusedFilter = query.tokens.find(
                        (token): token is Filter =>
                            // Inclusive end so that the filter is highlighted when
                            // the cursor is positioned directly after the value
                            token.type === 'filter' && token.range.start <= position && token.range.end >= position
                    )
                    const decorations: Range<Decoration>[] = []
                    if (focusedFilter) {
                        decorations.push(focusedFilterDeco.range(focusedFilter.range.start, focusedFilter.range.end))

                        if (!focusedFilter.value?.value) {
                            const resolvedFilter = resolveFilter(focusedFilter.field.value)
                            if (resolvedFilter?.definition.placeholder) {
                                decorations.push(
                                    Decoration.widget({
                                        widget: new PlaceholderWidget(resolvedFilter.definition.placeholder),
                                        side: 1, // show after the cursor
                                    }).range(focusedFilter.range.end)
                                )
                            }
                        }
                    }
                    this.decorations = Decoration.set(decorations)
                } else {
                    // Ideally we would remove the decoration when the input
                    // looses focus, but there is a bug that re-focuses the
                    // input when a filter is highlighted and the menu is
                    // clicked.
                    // See https://github.com/sourcegraph/sourcegraph/issues/38677
                    // this.decorations = Decoration.none
                }
            }
        },
    }),
    {
        decorations: plugin => plugin.decorations,
    }
)

// Extension for providing token information. This includes showing a popover on
// hover and highlighting the hovered token.
function tokenInfo(): Extension {
    const setHighlighedTokenPosition = StateEffect.define<number | null>()
    const highlightedTokenPosition = StateField.define<number | null>({
        create() {
            return null
        },
        update(position, transaction) {
            // Hide the highlight when the document changes. This replicates
            // Monaco's behavior.
            if (transaction.docChanged) {
                return null
            }
            const effect = transaction.effects.find((effect): effect is StateEffect<number | null> =>
                effect.is(setHighlighedTokenPosition)
            )
            if (effect) {
                position = effect?.value
            }
            if (position !== null) {
                // Mapping the position might not be necessary since we clear
                // the highlight when the document changes anyway, but this is
                // the safer way.
                // MapMode.TrackDel causes mapPos to return null if content at
                // this position was deleted (in which case we want to remove
                // the highlight)
                return transaction.changes.mapPos(position, 0, MapMode.TrackDel)
            }
            return position
        },
        provide(field) {
            return EditorView.decorations.compute([field, decoratedTokens], state => {
                const position = state.field(field)
                if (position === null) {
                    return Decoration.none
                }
                let tokenAtPosition = state.facet(decoratedTokens).find(token => isTokenInRange(position, token))

                switch (tokenAtPosition?.type) {
                    case 'field':
                    case 'pattern':
                    case 'metaRevision':
                    case 'metaRepoRevisionSeparator':
                    case 'metaSelector':
                    case 'metaRegexp':
                    case 'metaStructural':
                    case 'metaPredicate':
                        // These are the tokens we show hover information for
                        break
                    case 'keyword': {
                        // Find operator (AND and OR are supported)
                        const operator = findOperatorNode(position, state.facet(parsedQuery))
                        if (operator) {
                            return Decoration.set([
                                focusedFilterDeco.range(
                                    (operator.groupRange ?? operator.range).start,
                                    (operator.groupRange ?? operator.range).end
                                ),
                            ])
                        }
                        // Highlight operator keyword only
                        break
                    }

                    default:
                        tokenAtPosition = undefined
                        break
                }
                return tokenAtPosition
                    ? Decoration.set([
                          focusedFilterDeco.range(tokenAtPosition.range.start, getEndPosition(tokenAtPosition)),
                      ])
                    : Decoration.none
            })
        },
    })

    return [
        highlightedTokenPosition,
        // Highlights the hovered token
        EditorView.domEventHandlers({
            mousemove(event, view) {
                const position = view.posAtCoords(event)
                if (position && position !== view.state.field(highlightedTokenPosition)) {
                    view.dispatch({ effects: setHighlighedTokenPosition.of(position) })
                }
            },
            mouseleave(_event, view) {
                if (view.state.field(highlightedTokenPosition) !== null) {
                    view.dispatch({ effects: setHighlighedTokenPosition.of(null) })
                }
            },
        }),
        // Shows information about the hovered token
        hoverTooltip(
            (view, position) => {
                const tooltipInfo = getTokensTooltipInformation(view.state.facet(decoratedTokens), position)
                if (!tooltipInfo) {
                    return null
                }

                return {
                    pos: tooltipInfo.range.start,
                    end: tooltipInfo.range.end,
                    create(): TooltipView {
                        const dom = document.createElement('div')
                        dom.innerHTML = renderMarkdown(tooltipInfo.value)
                        return {
                            dom,
                        }
                    },
                }
            },
            {
                hoverTime: 100,
                // Hiding the tooltip when the document changes replicates
                // Monaco's behavior and also "feels right" because it removes
                // "clutter" from the input.
                hideOnChange: true,
            }
        ),
    ]
}

function getTokensTooltipInformation(
    tokens: DecoratedToken[],
    position: number
): { range: { start: number; end: number }; value: string } | null {
    const tokensAtCursor = tokens.filter(token => isTokenInRange(position, token))

    if (tokensAtCursor?.length === 0) {
        return null
    }
    const values: string[] = []
    let range: { start: number; end: number } | undefined

    // Copied and adapted from getHoverResult (hover.ts)
    for (const token of tokensAtCursor) {
        switch (token.type) {
            case 'field': {
                const resolvedFilter = resolveFilter(token.value)
                if (resolvedFilter) {
                    values.push(
                        'negated' in resolvedFilter
                            ? resolvedFilter.definition.description(resolvedFilter.negated)
                            : resolvedFilter.definition.description
                    )
                    range = { start: token.range.start, end: getEndPosition(token) }
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
            case 'keyword':
                switch (token.kind) {
                    case KeywordKind.Or:
                        values.push('Find results which match both the left and the right expression.')
                        range = token.range
                        break
                    case KeywordKind.And:
                        values.push('Find results which match the left or the right expression.')
                        range = token.range
                        break
                }
        }
    }

    if (!range) {
        return null
    }
    return { range, value: values.join('') }
}

// Hooks query diagnostics into the editor.
// The facet stores the diagnostics data which is used by the text decoration
// and the tooltip extensions.
const diagnostics = Facet.define<Monaco.IMarkerData[], Monaco.IMarkerData[]>({
    combine: markerData => markerData.flat(),
})
const diagnosticDecos: { [key in MarkerSeverity]: Decoration } = {
    [MarkerSeverity.Hint]: emptyDecorator,
    [MarkerSeverity.Info]: emptyDecorator,
    [MarkerSeverity.Warning]: Decoration.mark({ class: styles.diagnosticWarning }),
    [MarkerSeverity.Error]: Decoration.mark({ class: styles.diagnosticError }),
}
const queryDiagnostic: Extension = [
    // Compute diagnostics when query changes
    diagnostics.compute([queryTokens], state => {
        const query = state.facet(queryTokens)
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
        {
            hoverTime: 100,
            // Making changes elsewhere in the query might invalidate a specific
            // diagnostic (e.g. adding type:commit to a query that contains
            // author:...), so generally hiding them on any change seems
            // reasonable.
            hideOnChange: true,
        }
    ),
]

function findOperatorNode(position: number, node: Node | null): Extract<Node, { type: 'operator' }> | null {
    if (!node || node.type !== 'operator' || node.range.start >= position || node.range.end <= position) {
        return null
    }
    for (const operand of node.operands) {
        const result = findOperatorNode(position, operand)
        if (result) {
            return result
        }
    }
    return node
}

function isTokenInRange(position: number, token: Pick<DecoratedToken, 'type' | 'range'>): boolean {
    return token.range.start <= position && getEndPosition(token) > position
}

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function getEndPosition(token: Pick<DecoratedToken, 'type' | 'range'>): number {
    return token.range.end + (token.type === 'field' ? 1 : 0)
}
