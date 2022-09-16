import { basename } from 'path'

import {
    autocompletion,
    startCompletion,
    completionKeymap,
    CompletionResult,
    Completion,
    snippet,
    CompletionSource,
    acceptCompletion,
    selectedCompletion,
    currentCompletions,
    setSelectedCompletion,
} from '@codemirror/autocomplete'
import { Extension, Prec } from '@codemirror/state'
import { keymap, EditorView } from '@codemirror/view'
import {
    mdiCodeArray,
    mdiCodeBraces,
    mdiCodeNotEqual,
    mdiCodeString,
    mdiCube,
    mdiCubeOutline,
    mdiDrawingBox,
    mdiFileDocument,
    mdiFilterOutline,
    mdiFunction,
    mdiHistory,
    mdiKey,
    mdiLink,
    mdiMatrix,
    mdiNull,
    mdiNumeric,
    mdiPackage,
    mdiPiBox,
    mdiPillar,
    mdiPound,
    mdiShape,
    mdiSitemap,
    mdiSourceBranch,
    mdiTextBox,
    mdiTimetable,
    mdiWeb,
    mdiWrench,
} from '@mdi/js'
import { startCase } from 'lodash'

import { isDefined } from '@sourcegraph/common'
import { SymbolKind } from '@sourcegraph/search'
import {
    createFilterSuggestions,
    PREDICATE_REGEX,
    regexInsertText,
    repositoryInsertText,
} from '@sourcegraph/shared/src/search/query/completion'
import { decorate, DecoratedToken, toDecoration } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { getSuggestionQuery } from '@sourcegraph/shared/src/search/query/providers'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { queryTokens } from './parsedQuery'

type CompletionType = SymbolKind | 'queryfilter' | 'repository' | 'searchhistory'

// See SymbolIcon
const typeIconMap: Record<CompletionType, string> = {
    ACCELERATORS: mdiShape,
    ACCESSOR: mdiShape,
    ACTIVEBINDINGFUNC: mdiShape,
    ALIAS: mdiShape,
    ALTSTEP: mdiShape,
    ANCHOR: mdiShape,
    ANNOTATION: mdiCube,
    ANON: mdiShape,
    ANONMEMBER: mdiShape,
    ANTFILE: mdiShape,
    ARCHITECTURE: mdiShape,
    ARG: mdiShape,
    ARRAY: mdiCodeArray,
    ARTICLE: mdiShape,
    ARTIFACTID: mdiShape,
    ASSEMBLY: mdiShape,
    ASSERT: mdiShape,
    ATTRIBUTE: mdiShape,
    AUGROUP: mdiShape,
    AUTOVAR: mdiShape,
    BENCHMARK: mdiShape,
    BIBITEM: mdiShape,
    BITMAP: mdiShape,
    BLOCK: mdiShape,
    BLOCKDATA: mdiShape,
    BOOK: mdiShape,
    BOOKLET: mdiShape,
    BOOLEAN: mdiMatrix,
    BUILD: mdiShape,
    CALLBACK: mdiShape,
    CATEGORY: mdiShape,
    CCFLAG: mdiShape,
    CELL: mdiShape,
    CHAPTER: mdiShape,
    CHECKER: mdiShape,
    CHOICE: mdiShape,
    CHUNKLABEL: mdiShape,
    CITATION: mdiShape,
    CLASS: mdiSitemap,
    CLOCKING: mdiShape,
    COMBO: mdiShape,
    COMMAND: mdiShape,
    COMMON: mdiShape,
    COMPONENT: mdiShape,
    COND: mdiShape,
    CONDITION: mdiShape,
    CONFERENCE: mdiShape,
    CONFIG: mdiShape,
    CONST: mdiShape,
    CONSTANT: mdiPiBox,
    CONSTRAINT: mdiShape,
    CONSTRUCTOR: mdiCubeOutline,
    CONTEXT: mdiShape,
    COUNTER: mdiShape,
    COVERGROUP: mdiShape,
    CURSOR: mdiShape,
    CUSTOM: mdiShape,
    DATA: mdiShape,
    DATABASE: mdiShape,
    DATAFRAME: mdiShape,
    DEF: mdiShape,
    DEFINE: mdiShape,
    DEFINITION: mdiShape,
    DELEGATE: mdiShape,
    DELETEDFILE: mdiShape,
    DERIVEDMODE: mdiShape,
    DESCRIBE: mdiShape,
    DIALOG: mdiShape,
    DIRECTORY: mdiShape,
    DIVISION: mdiShape,
    DOCUMENT: mdiShape,
    DOMAIN: mdiShape,
    EDESC: mdiShape,
    ELEMENT: mdiShape,
    ENTITY: mdiShape,
    ENTRY: mdiShape,
    ENTRYSPEC: mdiShape,
    ENUM: mdiNumeric,
    ENUMCONSTANT: mdiNumeric,
    ENUMERATOR: mdiShape,
    ENVIRONMENT: mdiShape,
    ERROR: mdiShape,
    EVENT: mdiTimetable,
    EXCEPTION: mdiShape,
    EXTERNVAR: mdiShape,
    FACE: mdiShape,
    FD: mdiShape,
    FEATURE: mdiShape,
    FIELD: mdiTextBox,
    FILE: mdiFileDocument,
    FILENAME: mdiShape,
    FONT: mdiShape,
    FOOTNOTE: mdiShape,
    FORMAL: mdiShape,
    FORMAT: mdiShape,
    FRAGMENT: mdiShape,
    FRAMESUBTITLE: mdiShape,
    FRAMETITLE: mdiShape,
    FUN: mdiShape,
    FUNC: mdiShape,
    FUNCTION: mdiFunction,
    FUNCTIONVAR: mdiShape,
    FUNCTOR: mdiShape,
    GEM: mdiShape,
    GENERATOR: mdiShape,
    GENERIC: mdiShape,
    GETTER: mdiShape,
    GLOBAL: mdiShape,
    GLOBALVAR: mdiShape,
    GRAMMAR: mdiShape,
    GROUP: mdiShape,
    GROUPID: mdiShape,
    GUARD: mdiShape,
    HANDLER: mdiShape,
    HEADER: mdiShape,
    HEADING1: mdiShape,
    HEADING2: mdiShape,
    HEADING3: mdiShape,
    HEREDOC: mdiShape,
    HUNK: mdiShape,
    ICON: mdiShape,
    ID: mdiShape,
    IDENTIFIER: mdiShape,
    IFCLASS: mdiShape,
    IMPLEMENTATION: mdiShape,
    IMPORT: mdiShape,
    INBOOK: mdiShape,
    INCOLLECTION: mdiShape,
    INDEX: mdiShape,
    INFOITEM: mdiShape,
    INLINE: mdiShape,
    INPROCEEDINGS: mdiShape,
    INPUTSECTION: mdiShape,
    INSTANCE: mdiShape,
    INTEGER: mdiShape,
    INTERFACE: mdiLink,
    IPARAM: mdiShape,
    IT: mdiShape,
    KCONFIG: mdiShape,
    KEY: mdiKey,
    KEYWORD: mdiShape,
    KIND: mdiShape,
    L1HEADER: mdiShape,
    L2HEADER: mdiShape,
    L3HEADER: mdiShape,
    L4HEADER: mdiShape,
    L4SUBSECTION: mdiShape,
    L5HEADER: mdiShape,
    L5SUBSECTION: mdiShape,
    L6HEADER: mdiShape,
    LABEL: mdiShape,
    LANGDEF: mdiShape,
    LANGSTR: mdiShape,
    LIBRARY: mdiShape,
    LIST: mdiShape,
    LITERAL: mdiShape,
    LOCAL: mdiShape,
    LOCALVAR: mdiShape,
    LOCALVARIABLE: mdiShape,
    LOGGERSECTION: mdiShape,
    LTLIBRARY: mdiShape,
    MACRO: mdiShape,
    MACROFILE: mdiShape,
    MACROPARAM: mdiShape,
    MACROPARAMETER: mdiShape,
    MAINMENU: mdiShape,
    MAKEFILE: mdiShape,
    MAN: mdiShape,
    MANUAL: mdiShape,
    MAP: mdiShape,
    MASTERSTHESIS: mdiShape,
    MATCHEDTEMPLATE: mdiShape,
    MEMBER: mdiShape,
    MENU: mdiShape,
    MESSAGE: mdiShape,
    METHOD: mdiCubeOutline,
    METHODSPEC: mdiShape,
    MINORMODE: mdiShape,
    MISC: mdiShape,
    MIXIN: mdiShape,
    MLCONN: mdiShape,
    MLPROP: mdiShape,
    MLTABLE: mdiShape,
    MODIFIEDFILE: mdiShape,
    MODPORT: mdiShape,
    MODULE: mdiCodeBraces,
    MODULEPAR: mdiShape,
    MULTITASK: mdiShape,
    MXTAG: mdiShape,
    NAME: mdiShape,
    NAMEATTR: mdiShape,
    NAMEDPATTERN: mdiShape,
    NAMEDTEMPLATE: mdiShape,
    NAMELIST: mdiShape,
    NAMESPACE: mdiWeb,
    NET: mdiShape,
    NETTYPE: mdiShape,
    NEWFILE: mdiShape,
    NODE: mdiShape,
    NOTATION: mdiShape,
    NSPREFIX: mdiShape,
    NULL: mdiNull,
    NUMBER: mdiPound,
    OBJECT: mdiDrawingBox,
    ONEOF: mdiShape,
    OPARAM: mdiShape,
    OPERATOR: mdiCodeNotEqual,
    OPTENABLE: mdiShape,
    OPTION: mdiShape,
    OPTWITH: mdiShape,
    PACKAGE: mdiPackage,
    PACKAGENAME: mdiShape,
    PACKSPEC: mdiShape,
    PARAGRAPH: mdiShape,
    PARAM: mdiShape,
    PARAMETER: mdiCube,
    PARAMETERENTITY: mdiShape,
    PART: mdiShape,
    PATCH: mdiShape,
    PATH: mdiShape,
    PATTERN: mdiShape,
    PHANDLER: mdiShape,
    PHDTHESIS: mdiShape,
    PKG: mdiShape,
    PLACEHOLDER: mdiShape,
    PLAY: mdiShape,
    PORT: mdiShape,
    PROBE: mdiShape,
    PROCEDURE: mdiShape,
    PROCEEDINGS: mdiShape,
    PROCESS: mdiShape,
    PROGRAM: mdiShape,
    PROJECT: mdiShape,
    PROPERTY: mdiWrench,
    PROTECTED: mdiShape,
    PROTECTSPEC: mdiShape,
    PROTOCOL: mdiShape,
    PROTODEF: mdiShape,
    PROTOTYPE: mdiShape,
    PUBLICATION: mdiShape,
    QMP: mdiShape,
    QUALNAME: mdiShape,
    RECEIVER: mdiShape,
    RECORD: mdiShape,
    RECORDFIELD: mdiShape,
    REGEX: mdiShape,
    REGION: mdiShape,
    REGISTER: mdiShape,
    REOPEN: mdiShape,
    REPOID: mdiShape,
    REPOSITORYID: mdiShape,
    REPR: mdiShape,
    RESOURCE: mdiShape,
    RESPONSE: mdiShape,
    ROLE: mdiShape,
    ROOT: mdiShape,
    RPC: mdiShape,
    RULE: mdiShape,
    RUN: mdiShape,
    SCHEMA: mdiShape,
    SCRIPT: mdiShape,
    SECTION: mdiShape,
    SECTIONGROUP: mdiShape,
    SELECTOR: mdiShape,
    SEQUENCE: mdiShape,
    SERVER: mdiShape,
    SERVICE: mdiShape,
    SET: mdiShape,
    SETTER: mdiShape,
    SIGNAL: mdiShape,
    SIGNATURE: mdiShape,
    SINGLETONMETHOD: mdiShape,
    SLOT: mdiShape,
    SOURCE: mdiShape,
    SOURCEFILE: mdiShape,
    STEP: mdiShape,
    STRING: mdiCodeString,
    STRUCT: mdiPillar,
    STRUCTURE: mdiShape,
    STYLESHEET: mdiShape,
    SUBDIR: mdiShape,
    SUBMETHOD: mdiShape,
    SUBMODULE: mdiShape,
    SUBPARAGRAPH: mdiShape,
    SUBPROGRAM: mdiShape,
    SUBPROGSPEC: mdiShape,
    SUBROUTINE: mdiShape,
    SUBROUTINEDECLARATION: mdiShape,
    SUBSECTION: mdiShape,
    SUBSPEC: mdiShape,
    SUBST: mdiShape,
    SUBSTDEF: mdiShape,
    SUBSUBSECTION: mdiShape,
    SUBTITLE: mdiShape,
    SUBTYPE: mdiShape,
    SYMBOL: mdiShape,
    SYNONYM: mdiShape,
    TABLE: mdiShape,
    TAG: mdiShape,
    TALIAS: mdiShape,
    TARGET: mdiShape,
    TASK: mdiShape,
    TASKSPEC: mdiShape,
    TECHREPORT: mdiShape,
    TEMPLATE: mdiShape,
    TEST: mdiShape,
    TESTCASE: mdiShape,
    THEME: mdiShape,
    THEOREM: mdiShape,
    THRIFTFILE: mdiShape,
    THROWSPARAM: mdiShape,
    TIMER: mdiShape,
    TITLE: mdiShape,
    TOKEN: mdiShape,
    TOPLEVELVARIABLE: mdiShape,
    TPARAM: mdiShape,
    TRAIT: mdiShape,
    TRIGGER: mdiShape,
    TYPE: mdiShape,
    TYPEALIAS: mdiShape,
    TYPEDEF: mdiShape,
    TYPESPEC: mdiShape,
    UNION: mdiShape,
    UNIT: mdiShape,
    UNKNOWN: mdiShape,
    UNPUBLISHED: mdiShape,
    USERNAME: mdiShape,
    USING: mdiShape,
    VAL: mdiShape,
    VALUE: mdiShape,
    VAR: mdiShape,
    VARALIAS: mdiShape,
    VARIABLE: mdiCube,
    VARSPEC: mdiShape,
    VECTOR: mdiShape,
    VERSION: mdiShape,
    VIEW: mdiShape,
    VIRTUAL: mdiShape,
    VRESOURCE: mdiShape,
    WRAPPER: mdiShape,
    XINPUT: mdiShape,
    XTASK: mdiShape,
    queryfilter: mdiFilterOutline,
    repository: mdiSourceBranch,
    searchhistory: mdiHistory,
}

function createIcon(pathSpec: string): Node {
    const svgNS = 'http://www.w3.org/2000/svg'
    const svg = document.createElementNS(svgNS, 'svg')
    svg.setAttributeNS(null, 'viewBox', '0 0 24 24')
    svg.setAttribute('aria-hidden', 'true')

    const path = document.createElementNS(svgNS, 'path')
    path.setAttribute('d', pathSpec)

    svg.append(path)
    return svg
}

interface SuggestionContext {
    position: number
    onAbort: (listener: () => void) => void
}

/**
 * A suggestion source is given a completion context, the current tokens in the
 * query and the token at the current cursor position. It returns the
 * corresponding completion results.
 * The return type is generic so that it can be used to create different
 * suggestion structures.
 */
type SuggestionSource<R, C extends SuggestionContext> = (
    context: C,
    tokens: Token[],
    tokenAtPosition?: Token
) => R | null | Promise<R | null>

export type StandardSuggestionSource = SuggestionSource<CompletionResult | null, SuggestionContext>

/**
 * searchQueryAutocompletion registers extensions for automcompletion, using the
 * provided suggestion sources.
 */
export function searchQueryAutocompletion(
    sources: StandardSuggestionSource[],
    // By default we do not enable suggestion selection with enter because that
    // interferes with the query submission logic.
    applyOnEnter = false
): Extension {
    const override: CompletionSource[] = sources.map(source => context => {
        const position = context.pos
        const query = context.state.facet(queryTokens)
        const token = query.tokens.find(token => isTokenInRange(token, position))
        return source(
            { position, onAbort: listener => context.addEventListener('abort', listener) },
            query.tokens,
            token
        )
    })

    // Customizing how completion items are rendered
    const addToOptions: NonNullable<Parameters<typeof autocompletion>[0]>['addToOptions'] = [
        // This renders the completion icon
        {
            render(completion) {
                return createIcon(
                    completion.type && completion.type in typeIconMap
                        ? typeIconMap[completion.type as CompletionType]
                        : typeIconMap[SymbolKind.UNKNOWN]
                )
            },
            // Per CodeMirror documentation, 20 is the default icon
            // position
            position: 20,
        },
        {
            render(completion) {
                if (completion.type !== 'searchhistory') {
                    return null
                }
                const tokens = scanSearchQuery(completion.label)
                if (tokens.type !== 'success') {
                    throw new Error('this should not happen')
                }
                const nodes = tokens.term
                    .flatMap(token => decorate(token))
                    .map(token => {
                        const decoration = toDecoration(completion.label, token)
                        const node = document.createElement('span')
                        node.className = decoration.className
                        node.textContent = decoration.value
                        return node
                    })

                const container = document.createElement('div')
                container.style.whiteSpace = 'initial'
                for (const node of nodes) {
                    container.append(node)
                }
                return container
            },
            position: 30,
        },
    ]

    return [
        // Uses the default keymapping but changes accepting suggestions from Enter
        // to Tab
        Prec.highest(
            keymap.of(
                applyOnEnter
                    ? [
                          ...completionKeymap,
                          {
                              key: 'Tab',
                              run(view) {
                                  // Select first completion item if none is selected
                                  // and items are available.
                                  if (selectedCompletion(view.state) === null) {
                                      if (currentCompletions(view.state).length > 0) {
                                          view.dispatch({ effects: setSelectedCompletion(0) })
                                          return true
                                      }
                                      return false
                                  }
                                  // Otherwise apply the selected completion item
                                  return acceptCompletion(view)
                              },
                          },
                      ]
                    : completionKeymap.map(keybinding =>
                          keybinding.key === 'Enter' ? { ...keybinding, key: 'Tab' } : keybinding
                      )
            )
        ),
        EditorView.theme({
            '.completion-type-queryfilter > .cm-completionLabel': {
                fontWeight: 'bold',
            },
            '.cm-tooltip-autocomplete svg': {
                width: '1rem',
                height: '1rem',
                display: 'inline-block',
                boxSizing: 'content-box',
                textAlign: 'center',
                paddingRight: '0.5rem',
            },
            '.cm-tooltip-autocomplete svg path': {
                fillOpacity: 0.6,
            },
            '.completion-type-searchhistory > .cm-completionLabel': {
                display: 'none',
            },
            'li.completion-type-searchhistory': {
                height: 'initial !important',
                minHeight: '1.3rem',
            },
        }),
        EditorView.updateListener.of(update => {
            // If a filter was completed, show the completion list again for
            // filter values.
            if (update.transactions.some(transaction => transaction.isUserEvent('input.complete'))) {
                const query = update.state.facet(queryTokens)
                const token = query.tokens.find(token => isTokenInRange(token, update.state.selection.main.anchor - 1))
                if (token) {
                    startCompletion(update.view)
                }
            }
        }),
        autocompletion({
            // We define our own keymap above
            defaultKeymap: false,
            override,
            optionClass: completionItem => 'completion-type-' + (completionItem.type ?? ''),
            icons: false,
            closeOnBlur: true,
            selectOnOpen: !applyOnEnter,
            addToOptions,
        }),
    ]
}

export interface DefaultSuggestionSourcesOptions {
    fetchSuggestions: (query: string, onAbort: (listener: () => void) => void) => Promise<SearchMatch[]>
    isSourcegraphDotCom: boolean
    globbing: boolean
    disableFilterCompletion?: true
    disableSymbolCompletion?: true
    showWhenEmpty?: boolean
}

/**
 * Creates default suggestion sources to complete available filters, dynamic
 * suggestions for the current pattern and static and dynamic suggestions for
 * the current filter value.
 */
export function createDefaultSuggestionSources(
    options: DefaultSuggestionSourcesOptions
): SuggestionSource<CompletionResult | null, SuggestionContext>[] {
    const sources: SuggestionSource<CompletionResult | null, SuggestionContext>[] = []

    if (options.disableFilterCompletion !== true) {
        sources.push(
            // Static suggestions shown if the current position is outside a
            // filter value
            createDefaultSource((context, tokens, token) => {
                if (tokens.length === 0 && options.showWhenEmpty === false) {
                    return null
                }

                // Default to the current cursor position (e.g. if the token is a
                // whitespace, we want the suggestion to be inserted after it)
                let from = context.position

                if (token?.type === 'pattern') {
                    // If the token is a pattern (e.g. the start of a filter name),
                    // we want the suggestion to complete that name.
                    from = token.range.start
                }

                return {
                    from,
                    options: FILTER_SUGGESTIONS,
                }
            }),
            // Show static filter value suggestions
            createFilterSource((_context, _tokens, token, resolvedFilter) => {
                if (!resolvedFilter?.definition.discreteValues) {
                    return null
                }

                const { value } = token
                const insidePredicate = value ? PREDICATE_REGEX.test(value.value) : false
                const hasDynamicSuggestions = resolvedFilter.definition.suggestions

                // Don't show static suggestions if we are inside a predicate or
                // if the filter already has a value _and_ is configured for
                // dynamic suggestions.
                // That's because dynamic suggestions are not filtered (filter: false)
                // which CodeMirror always displays above filtered suggestions.
                if (insidePredicate || (value && hasDynamicSuggestions)) {
                    return null
                }

                return {
                    from: value?.range.start ?? token.range.end,
                    to: value?.range.end,
                    // Filtering is unnecessary when dynamic suggestions are
                    // available because if there is any input that the static
                    // suggestions could be filtered by we disable static
                    // suggestions and only show the dynamic ones anyway.
                    filter: !hasDynamicSuggestions,
                    options: resolvedFilter.definition
                        .discreteValues(value, options.isSourcegraphDotCom)
                        .map(({ label, insertText, asSnippet }, index) => {
                            const apply = (insertText || label) + ' '
                            return {
                                label,
                                // See issue https://github.com/sourcegraph/sourcegraph/issues/38254
                                // Per CodeMirror's documentation (https://codemirror.net/docs/ref/#autocomplete.snippet)
                                // "The user can move between fields with Tab and Shift-Tab as long as the fields are
                                // active. Moving to the last field or moving the cursor out of the current field
                                // deactivates the fields."
                                // This means we need to append a field at the end so that pressing Tab when at the last
                                // field will move the cursor after the filter value and not move focus outside the input
                                apply: asSnippet ? snippet(apply + '${}') : apply,
                                // Setting boost this way has the effect of
                                // displaying matching suggestions in the same
                                // order as they have been defined in code.
                                boost: index * -1,
                            }
                        }),
                }
            }),

            // Show dynamic filter value suggestions
            createFilterSource(async (context, tokens, token, resolvedFilter) => {
                // On Sourcegraph.com, prompt only static suggestions (above) if there is no value to use for generating dynamic suggestions yet.
                if (
                    options.isSourcegraphDotCom &&
                    (!token.value || (token.value.type === 'literal' && token.value.value === ''))
                ) {
                    return null
                }

                if (!resolvedFilter?.definition.suggestions) {
                    return null
                }

                const results = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, resolvedFilter.definition.suggestions),
                    context.onAbort
                )
                if (results.length === 0) {
                    return null
                }
                const filteredResults = results
                    .filter(match => match.type === resolvedFilter.definition.suggestions)
                    .map(match => {
                        switch (match.type) {
                            case 'path':
                                return {
                                    label: match.path,
                                    type: SymbolKind.FILE,
                                    apply: regexInsertText(match.path, options) + ' ',
                                    info: match.repository,
                                }
                            case 'repo':
                                return {
                                    label: match.repository,
                                    type: 'repository',
                                    apply:
                                        repositoryInsertText(match, { ...options, filterValue: token.value?.value }) +
                                        ' ',
                                }
                        }
                        return null
                    })
                    .filter(isDefined)

                const insidePredicate = token.value ? PREDICATE_REGEX.test(token.value.value) : false

                return {
                    from: token.value?.range.start ?? token.range.end,
                    to: token.value?.range.end,
                    filter: false,
                    options: filteredResults,
                    getMatch: insidePredicate || options.globbing ? undefined : createMatchFunction(token),
                }
            })
        )
    }

    if (options.disableSymbolCompletion !== true) {
        sources.push(
            // Show symbol suggestions outside of filters
            createDefaultSource(async (context, tokens, token) => {
                if (!token || token.type !== 'pattern') {
                    return null
                }

                const results = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, 'symbol'),
                    context.onAbort
                )
                if (results.length === 0) {
                    return null
                }

                return {
                    from: token.range.start,
                    to: token.range.end,
                    options: results
                        .flatMap(result => {
                            if (result.type === 'symbol') {
                                const path = result.path
                                return result.symbols.map(symbol => ({
                                    label: symbol.name,
                                    type: symbol.kind,
                                    apply: symbol.name + ' ',
                                    detail: `${startCase(symbol.kind.toLowerCase())} | ${basename(path)}`,
                                    info: result.repository,
                                }))
                            }
                            return null
                        })
                        .filter(isDefined),
                }
            })
        )
    }

    return sources
}

/**
 * Creates a suggestion source that triggers on no token or pattern or whitespace
 * tokens.
 */
function createDefaultSource<R, C extends SuggestionContext>(source: SuggestionSource<R, C>): SuggestionSource<R, C> {
    return (context, tokens, token) => {
        if (token && token.type !== 'pattern' && token.type !== 'whitespace') {
            return null
        }
        return source(context, tokens, token)
    }
}

type FilterSuggestionSource<R, C extends SuggestionContext> = (
    context: C,
    tokens: Token[],
    filter: Filter,
    resolvedFilter: ReturnType<typeof resolveFilter>
) => ReturnType<SuggestionSource<R, C>>

/**
 * Creates a suggestion source that triggers when a filter value is completed.
 */
function createFilterSource<R, C extends SuggestionContext>(
    source: FilterSuggestionSource<R, C>
): SuggestionSource<R, C> {
    return (context, tokens, token) => {
        // Not completing filter value
        if (!token || token.type !== 'filter' || (token.value && token.value.range.start > context.position)) {
            return null
        }

        const resolvedFilter = resolveFilter(token.field.value)
        if (!resolvedFilter) {
            return null
        }

        return source(context, tokens, token, resolvedFilter)
    }
}

const FILTER_SUGGESTIONS: Completion[] = createFilterSuggestions(Object.keys(FILTERS) as FilterType[]).map(
    ({ label, insertText, detail }) => ({
        label,
        type: 'queryfilter',
        apply: insertText,
        detail,
        boost: insertText.startsWith('-') ? 1 : 2, // demote negated filters
    })
)

/**
 * This helper function creates a function suitable for CodeMirror's 'getMatch'
 * option. This is used to allow CodeMirror to highlight the matching part of
 * the label.
 * See https://codemirror.net/docs/ref/#autocomplete.CompletionResult.getMatch
 */
function createMatchFunction(token: Filter): ((completion: Completion) => number[]) | undefined {
    if (!token.value?.value) {
        return undefined
    }
    try {
        // Creating a regular expression fails if the value contains special
        // regex characters in invalid positions. In that case we don't
        // highlight.
        const pattern = new RegExp(token.value.value, 'ig')
        return completion => Array.from(completion.label.matchAll(pattern), matchToIndexTuple).flat()
    } catch {
        return undefined
    }
}

/**
 * Converts a regular expression match into an (possibly empty) number tuple
 * representing the start index and the end index of the match.
 */
function matchToIndexTuple(match: RegExpMatchArray): number[] {
    return match.index !== undefined ? [match.index, match.index + match[0].length] : []
}

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function isTokenInRange(
    token: { type: DecoratedToken['type']; range: { start: number; end: number } },
    position: number
): boolean {
    return token.range.start <= position && token.range.end + (token.type === 'field' ? 2 : 0) >= position
}
