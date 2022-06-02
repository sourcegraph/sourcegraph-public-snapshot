import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { startCompletion } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import {
    EditorSelection,
    EditorState,
    Extension,
    Facet,
    StateEffect,
    StateField,
    Prec,
    RangeSetBuilder,
    MapMode,
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
} from '@codemirror/view'
import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import { editor as Monaco, MarkerSeverity } from 'monaco-editor'

import { renderMarkdown } from '@sourcegraph/common'
import { QueryChangeSource, SearchPatternTypeProps } from '@sourcegraph/search'
import { useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'
import { resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { toHover } from '@sourcegraph/shared/src/search/query/hover'
import { createCancelableFetchSuggestions } from '@sourcegraph/shared/src/search/query/providers'
import { Filter } from '@sourcegraph/shared/src/search/query/token'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { searchQueryAutocompletion, createDefaultSuggestionSources } from './extensions/completion'
import { decoratedTokens, parsedQuery, parseInputAsQuery, setQueryParseOptions } from './extensions/parsedQuery'
import { MonacoQueryInputProps } from './MonacoQueryInput'

import styles from './CodeMirrorQueryInput.module.scss'

const replacePattern = /[\n\r↵]/g

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
}) => {
    const value = preventNewLine ? queryState.query.replace(replacePattern, '') : queryState.query
    // We use both, state and a ref, for the editor instance because we need to
    // re-run some hooks when the editor changes but we also need a stable
    // reference that doesn't change across renders (and some hooks should only
    // run when a prop changes, not the editor).
    const [editor, setEditor] = useState<EditorView | undefined>()
    const editorReference = useRef<EditorView>()

    const hasSubmitHandler = onSubmit !== undefined

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
            searchQueryAutocompletion(
                createDefaultSuggestionSources({
                    fetchSuggestions: createCancelableFetchSuggestions(query =>
                        fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpec))
                    ),
                    globbing,
                    isSourcegraphDotCom,
                })
            ),
        [selectedSearchContextSpec, globbing, isSourcegraphDotCom]
    )

    const extensions = useMemo(() => {
        const extensions: Extension[] = [
            EditorView.contentAttributes.of({ 'aria-label': ariaLabel }),
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

        if (hasSubmitHandler) {
            extensions.push(Prec.high(notifyOnEnter))
        }

        if (onHandleFuzzyFinder) {
            extensions.push(keymap.of([{ key: 'Mod-k', run: () => (onHandleFuzzyFinder(true), true) }]))
        }

        if (preventNewLine) {
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
    }, [
        ariaLabel,
        autocompletion,
        onBlur,
        onChange,
        onHandleFuzzyFinder,
        hasSubmitHandler,
        placeholder,
        preventNewLine,
        editorOptions,
    ])

    // We use an effect + field to configure the submission handler so that we
    // don't reconfigure the whole editor should the 'onSubmit' handler change
    // because of the changed query.
    useEffect(() => {
        if (editor && onSubmit) {
            editor.dispatch({ effects: [setNotifyHandler.of(onSubmit)] })
        }
    }, [editor, onSubmit])

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

        switch (queryState.changeSource) {
            case QueryChangeSource.userInput:
                // Don't react to user input
                break
            case QueryChangeSource.searchTypes:
            case QueryChangeSource.searchReference: {
                // Select the specified range (most of the time this will be a
                // placeholder filter value).
                const selectionRange = queryState.selectionRange
                editor.dispatch({
                    selection: EditorSelection.range(selectionRange.start, selectionRange.end),
                    scrollIntoView: true,
                })
                editor.focus()
                if (queryState.showSuggestions) {
                    startCompletion(editor)
                }
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
const CodeMirrorQueryInput: React.FunctionComponent<React.PropsWithChildren<CodeMirrorQueryInputProps>> = React.memo(
    ({ isLightTheme, onEditorCreated, patternType, interpretComments, value, className, extensions = [] }) => {
        // This is using state instead of a ref because `useRef` doesn't cause a
        // re-render when the ref is attached, but we need that so that
        // `useCodeMirror` is called again and the editor is actually created.
        const [container, setContainer] = useState<HTMLDivElement | null>(null)

        const editor = useCodeMirror(
            container,
            value,
            useMemo(
                () => [
                    keymap.of(historyKeymap),
                    keymap.of(defaultKeymap),
                    history(),
                    EditorView.darkTheme.of(isLightTheme === false),
                    parseInputAsQuery({ patternType, interpretComments }),
                    tokenHighlight,
                    queryDiagnostic,
                    tokenInfo(),
                    highlightFocusedFilter,
                    ...extensions,
                ],
                // patternType and interpretComments are updated via a
                // transaction since there is no need to re-initialize all
                // extensions
                // eslint-disable-next-line react-hooks/exhaustive-deps
                [isLightTheme, extensions]
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
            editor?.dispatch({ effects: [setQueryParseOptions.of({ patternType, interpretComments })] })
        }, [editor, patternType, interpretComments])

        return <div ref={setContainer} className={classNames(styles.root, className)} id="monaco-query-input" />
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

// Enforces that the input won't split over multiple lines (basically prevents
// Enter from inserting a new line)
const singleLine = EditorState.transactionFilter.of(transaction => (transaction.newDoc.lines > 1 ? [] : transaction))

// Binds a function to the Enter key. Instead of using keymap directly, this is
// configured via a state field that contains the event handler. This way the
// event handler can be updated without having to reconfigure the whole editor.
// The event handler must be set via the setNotifyHandler effect.
const setNotifyHandler = StateEffect.define<() => void>()
const notifyOnEnter = StateField.define<() => void>({
    create() {
        return () => {}
    },
    update(value, transaction) {
        const effect = transaction.effects.find((effect): effect is StateEffect<() => void> =>
            effect.is(setNotifyHandler)
        )
        return effect ? effect.value : value
    },
    provide(field) {
        return keymap.of([
            {
                key: 'Enter',
                run: view => {
                    view.state.field(field)?.()
                    return true
                },
            },
        ])
    },
})

// Defines decorators for syntax highlighting
type StyleNames = keyof typeof styles
const tokenDecorators: { [key: string]: Decoration } = Object.fromEntries(
    (Object.keys(styles) as StyleNames[]).map(style => [style, Decoration.mark({ class: styles[style] })])
)
const emptyDecorator = Decoration.mark({})
const focusedFilterDeco = Decoration.mark({ class: styles.focusedFilter })

// Chooses the correct decorator for the decorated token. Copied (and adapted)
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

// This provides syntax highlighting. This is a custom solution so that we an
// use our existing query parser (instead of using CodeMirror's language
// support). That's not to say that we couldn't properly integrate with
// CodeMirror's language system with more effort.
const tokenHighlight = EditorView.decorations.compute([decoratedTokens], state => {
    const tokens = state.facet(decoratedTokens)
    const builder = new RangeSetBuilder<Decoration>()
    for (const token of tokens) {
        builder.add(token.range.start, getEndPosition(token), decoratedToDecoration(token))
    }
    return builder.finish()
})

// Determines whether the cursor is over a filter and if yes, decorates that
// filter.
const highlightFocusedFilter = ViewPlugin.define(
    () => ({
        decorations: Decoration.none,
        update(update) {
            if (update.focusChanged && !update.view.hasFocus) {
                this.decorations = Decoration.none
            } else if (update.docChanged || update.selectionSet || update.focusChanged) {
                const query = update.state.facet(parsedQuery)
                const position = update.state.selection.main.head
                const focusedFilter = query.tokens.find(
                    (token): token is Filter =>
                        // Inclusive end so that the filter is highlighed when
                        // the cursor is positioned directly after the value
                        token.type === 'filter' && token.range.start <= position && token.range.end >= position
                )
                this.decorations = focusedFilter
                    ? Decoration.set(focusedFilterDeco.range(focusedFilter.range.start, focusedFilter.range.end))
                    : Decoration.none
            }
        },
    }),
    {
        decorations: plugin => plugin.decorations,
    }
)

// Tooltip information.
function tokenInfo(): Extension[] {
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
                    view.dispatch({ effects: [setHighlighedTokenPosition.of(position)] })
                }
            },
            mouseleave(_event, view) {
                if (view.state.field(highlightedTokenPosition) !== null) {
                    view.dispatch({ effects: [setHighlighedTokenPosition.of(null)] })
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
const queryDiagnostic: Extension[] = [
    // Compute diagnostics when query changes
    diagnostics.compute([parsedQuery], state => {
        const query = state.facet(parsedQuery)
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

function isTokenInRange(position: number, token: Pick<DecoratedToken, 'type' | 'range'>): boolean {
    return token.range.start <= position && getEndPosition(token) > position
}

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function getEndPosition(token: Pick<DecoratedToken, 'type' | 'range'>): number {
    return token.range.end + (token.type === 'field' ? 1 : 0)
}
