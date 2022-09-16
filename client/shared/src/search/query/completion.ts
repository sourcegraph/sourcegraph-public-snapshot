import { escapeRegExp, startCase } from 'lodash'
import * as Monaco from 'monaco-editor'
import { Omit } from 'utility-types'

import { SymbolKind } from '../../graphql-operations'
import { RepositoryMatch, SearchMatch } from '../stream'

import { FilterType, isNegatableFilter, resolveFilter, FILTERS, escapeSpaces, ResolvedFilter } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import { CharacterRange, Filter, Pattern, Token, Whitespace } from './token'

export const repositoryCompletionItemKind = Monaco.languages.CompletionItemKind.Color
const filterCompletionItemKind = Monaco.languages.CompletionItemKind.Issue

type PartialCompletionItem = Omit<Monaco.languages.CompletionItem, 'range'>

export const PREDICATE_REGEX = /^([.A-Za-z]+)\((.*?)\)?$/

/**
 * COMPLETION_ITEM_SELECTED is a custom Monaco command that we fire after the user selects an autocomplete suggestion.
 * This allows us to be notified and run custom code when a user selects a suggestion.
 */
export const COMPLETION_ITEM_SELECTED: Monaco.languages.Command = {
    id: 'completionItemSelected',
    title: 'completion item selected',
}

/**
 * Given a list of filter types, this function returns a list of objects which
 * can be used for creating completion items. The result also includes negated
 * entries for negateable filters.
 */
export const createFilterSuggestions = (
    filter: FilterType[]
): { label: string; insertText: string; filterText: string; detail: string }[] =>
    filter.flatMap(filterType => {
        const completionItem = {
            label: filterType,
            insertText: `${filterType}:`,
            filterText: filterType,
            detail: '',
        }
        if (isNegatableFilter(filterType)) {
            return [
                {
                    ...completionItem,
                    detail: FILTERS[filterType].description(false),
                },
                {
                    ...completionItem,
                    label: `-${filterType}`,
                    insertText: `-${filterType}:`,
                    filterText: `-${filterType}`,
                    detail: FILTERS[filterType].description(true),
                },
            ]
        }
        return [
            {
                ...completionItem,
                detail: FILTERS[filterType].description,
            },
        ]
    })

/**
 * Default filter completions for all filters.
 */
const FILTER_TYPE_COMPLETIONS: Omit<Monaco.languages.CompletionItem, 'range'>[] = createFilterSuggestions(
    Object.keys(FILTERS) as FilterType[]
)
    // Set a sortText so that filter type suggestions
    // are shown before dynamic suggestions.
    .map((completionItem, index) => ({
        ...completionItem,
        kind: filterCompletionItemKind,
        sortText: `0${index}`,
    }))

/**
 * regexInsertText escapes the provided value so that it can be used as value
 * for a filter which expects a regular expression.
 */
export const regexInsertText = (value: string, options: { globbing: boolean }): string => {
    const insertText = options.globbing ? value : `^${escapeRegExp(value)}$`
    return escapeSpaces(insertText)
}

/**
 * repositoryInsertText escapes the provides value so that it can be used as a
 * value for the `repo:` filter.
 */
export const repositoryInsertText = (
    { repository }: RepositoryMatch,
    options: { globbing: boolean; filterValue?: string }
): string => regexInsertText(repository, options)

/**
 * Maps Sourcegraph SymbolKinds to Monaco CompletionItemKinds.
 */
const symbolKindToCompletionItemKind: Record<SymbolKind, Monaco.languages.CompletionItemKind> = {
    ACCELERATORS: Monaco.languages.CompletionItemKind.Value,
    ACCESSOR: Monaco.languages.CompletionItemKind.Value,
    ACTIVEBINDINGFUNC: Monaco.languages.CompletionItemKind.Value,
    ALIAS: Monaco.languages.CompletionItemKind.Value,
    ALTSTEP: Monaco.languages.CompletionItemKind.Value,
    ANCHOR: Monaco.languages.CompletionItemKind.Value,
    ANNOTATION: Monaco.languages.CompletionItemKind.TypeParameter,
    ANON: Monaco.languages.CompletionItemKind.Value,
    ANONMEMBER: Monaco.languages.CompletionItemKind.Value,
    ANTFILE: Monaco.languages.CompletionItemKind.Value,
    ARCHITECTURE: Monaco.languages.CompletionItemKind.Value,
    ARG: Monaco.languages.CompletionItemKind.Value,
    ARRAY: Monaco.languages.CompletionItemKind.Value,
    ARTICLE: Monaco.languages.CompletionItemKind.Value,
    ARTIFACTID: Monaco.languages.CompletionItemKind.Value,
    ASSEMBLY: Monaco.languages.CompletionItemKind.Value,
    ASSERT: Monaco.languages.CompletionItemKind.Value,
    ATTRIBUTE: Monaco.languages.CompletionItemKind.Value,
    AUGROUP: Monaco.languages.CompletionItemKind.Value,
    AUTOVAR: Monaco.languages.CompletionItemKind.Value,
    BENCHMARK: Monaco.languages.CompletionItemKind.Value,
    BIBITEM: Monaco.languages.CompletionItemKind.Value,
    BITMAP: Monaco.languages.CompletionItemKind.Value,
    BLOCK: Monaco.languages.CompletionItemKind.Value,
    BLOCKDATA: Monaco.languages.CompletionItemKind.Value,
    BOOK: Monaco.languages.CompletionItemKind.Value,
    BOOKLET: Monaco.languages.CompletionItemKind.Value,
    BOOLEAN: Monaco.languages.CompletionItemKind.Value,
    BUILD: Monaco.languages.CompletionItemKind.Value,
    CALLBACK: Monaco.languages.CompletionItemKind.Value,
    CATEGORY: Monaco.languages.CompletionItemKind.Value,
    CCFLAG: Monaco.languages.CompletionItemKind.Value,
    CELL: Monaco.languages.CompletionItemKind.Value,
    CHAPTER: Monaco.languages.CompletionItemKind.Value,
    CHECKER: Monaco.languages.CompletionItemKind.Value,
    CHOICE: Monaco.languages.CompletionItemKind.Value,
    CHUNKLABEL: Monaco.languages.CompletionItemKind.Value,
    CITATION: Monaco.languages.CompletionItemKind.Value,
    CLASS: Monaco.languages.CompletionItemKind.Class,
    CLOCKING: Monaco.languages.CompletionItemKind.Value,
    COMBO: Monaco.languages.CompletionItemKind.Value,
    COMMAND: Monaco.languages.CompletionItemKind.Value,
    COMMON: Monaco.languages.CompletionItemKind.Value,
    COMPONENT: Monaco.languages.CompletionItemKind.Value,
    COND: Monaco.languages.CompletionItemKind.Value,
    CONDITION: Monaco.languages.CompletionItemKind.Value,
    CONFERENCE: Monaco.languages.CompletionItemKind.Value,
    CONFIG: Monaco.languages.CompletionItemKind.Value,
    CONST: Monaco.languages.CompletionItemKind.Value,
    CONSTANT: Monaco.languages.CompletionItemKind.Constant,
    CONSTRAINT: Monaco.languages.CompletionItemKind.Value,
    CONSTRUCTOR: Monaco.languages.CompletionItemKind.Constructor,
    CONTEXT: Monaco.languages.CompletionItemKind.Value,
    COUNTER: Monaco.languages.CompletionItemKind.Value,
    COVERGROUP: Monaco.languages.CompletionItemKind.Value,
    CURSOR: Monaco.languages.CompletionItemKind.Value,
    CUSTOM: Monaco.languages.CompletionItemKind.Value,
    DATA: Monaco.languages.CompletionItemKind.Value,
    DATABASE: Monaco.languages.CompletionItemKind.Value,
    DATAFRAME: Monaco.languages.CompletionItemKind.Value,
    DEF: Monaco.languages.CompletionItemKind.Value,
    DEFINE: Monaco.languages.CompletionItemKind.Value,
    DEFINITION: Monaco.languages.CompletionItemKind.Value,
    DELEGATE: Monaco.languages.CompletionItemKind.Value,
    DELETEDFILE: Monaco.languages.CompletionItemKind.Value,
    DERIVEDMODE: Monaco.languages.CompletionItemKind.Value,
    DESCRIBE: Monaco.languages.CompletionItemKind.Value,
    DIALOG: Monaco.languages.CompletionItemKind.Value,
    DIRECTORY: Monaco.languages.CompletionItemKind.Value,
    DIVISION: Monaco.languages.CompletionItemKind.Value,
    DOCUMENT: Monaco.languages.CompletionItemKind.Value,
    DOMAIN: Monaco.languages.CompletionItemKind.Value,
    EDESC: Monaco.languages.CompletionItemKind.Value,
    ELEMENT: Monaco.languages.CompletionItemKind.Value,
    ENTITY: Monaco.languages.CompletionItemKind.Value,
    ENTRY: Monaco.languages.CompletionItemKind.Value,
    ENTRYSPEC: Monaco.languages.CompletionItemKind.Value,
    ENUM: Monaco.languages.CompletionItemKind.Enum,
    ENUMCONSTANT: Monaco.languages.CompletionItemKind.EnumMember,
    ENUMERATOR: Monaco.languages.CompletionItemKind.Value,
    ENVIRONMENT: Monaco.languages.CompletionItemKind.Value,
    ERROR: Monaco.languages.CompletionItemKind.Value,
    EVENT: Monaco.languages.CompletionItemKind.Event,
    EXCEPTION: Monaco.languages.CompletionItemKind.Value,
    EXTERNVAR: Monaco.languages.CompletionItemKind.Value,
    FACE: Monaco.languages.CompletionItemKind.Value,
    FD: Monaco.languages.CompletionItemKind.Value,
    FEATURE: Monaco.languages.CompletionItemKind.Value,
    FIELD: Monaco.languages.CompletionItemKind.Field,
    FILE: Monaco.languages.CompletionItemKind.File,
    FILENAME: Monaco.languages.CompletionItemKind.Value,
    FONT: Monaco.languages.CompletionItemKind.Value,
    FOOTNOTE: Monaco.languages.CompletionItemKind.Value,
    FORMAL: Monaco.languages.CompletionItemKind.Value,
    FORMAT: Monaco.languages.CompletionItemKind.Value,
    FRAGMENT: Monaco.languages.CompletionItemKind.Value,
    FRAMESUBTITLE: Monaco.languages.CompletionItemKind.Value,
    FRAMETITLE: Monaco.languages.CompletionItemKind.Value,
    FUN: Monaco.languages.CompletionItemKind.Function,
    FUNC: Monaco.languages.CompletionItemKind.Function,
    FUNCTION: Monaco.languages.CompletionItemKind.Function,
    FUNCTIONVAR: Monaco.languages.CompletionItemKind.Value,
    FUNCTOR: Monaco.languages.CompletionItemKind.Value,
    GEM: Monaco.languages.CompletionItemKind.Value,
    GENERATOR: Monaco.languages.CompletionItemKind.Value,
    GENERIC: Monaco.languages.CompletionItemKind.Value,
    GETTER: Monaco.languages.CompletionItemKind.Value,
    GLOBAL: Monaco.languages.CompletionItemKind.Value,
    GLOBALVAR: Monaco.languages.CompletionItemKind.Value,
    GRAMMAR: Monaco.languages.CompletionItemKind.Value,
    GROUP: Monaco.languages.CompletionItemKind.Value,
    GROUPID: Monaco.languages.CompletionItemKind.Value,
    GUARD: Monaco.languages.CompletionItemKind.Value,
    HANDLER: Monaco.languages.CompletionItemKind.Value,
    HEADER: Monaco.languages.CompletionItemKind.Value,
    HEADING1: Monaco.languages.CompletionItemKind.Value,
    HEADING2: Monaco.languages.CompletionItemKind.Value,
    HEADING3: Monaco.languages.CompletionItemKind.Value,
    HEREDOC: Monaco.languages.CompletionItemKind.Value,
    HUNK: Monaco.languages.CompletionItemKind.Value,
    ICON: Monaco.languages.CompletionItemKind.Value,
    ID: Monaco.languages.CompletionItemKind.Value,
    IDENTIFIER: Monaco.languages.CompletionItemKind.Value,
    IFCLASS: Monaco.languages.CompletionItemKind.Value,
    IMPLEMENTATION: Monaco.languages.CompletionItemKind.Value,
    IMPORT: Monaco.languages.CompletionItemKind.Value,
    INBOOK: Monaco.languages.CompletionItemKind.Value,
    INCOLLECTION: Monaco.languages.CompletionItemKind.Value,
    INDEX: Monaco.languages.CompletionItemKind.Value,
    INFOITEM: Monaco.languages.CompletionItemKind.Value,
    INLINE: Monaco.languages.CompletionItemKind.Value,
    INPROCEEDINGS: Monaco.languages.CompletionItemKind.Value,
    INPUTSECTION: Monaco.languages.CompletionItemKind.Value,
    INSTANCE: Monaco.languages.CompletionItemKind.Value,
    INTEGER: Monaco.languages.CompletionItemKind.Value,
    INTERFACE: Monaco.languages.CompletionItemKind.Interface,
    IPARAM: Monaco.languages.CompletionItemKind.Value,
    IT: Monaco.languages.CompletionItemKind.Value,
    KCONFIG: Monaco.languages.CompletionItemKind.Value,
    KEY: Monaco.languages.CompletionItemKind.Property,
    KEYWORD: Monaco.languages.CompletionItemKind.Value,
    KIND: Monaco.languages.CompletionItemKind.Value,
    L1HEADER: Monaco.languages.CompletionItemKind.Value,
    L2HEADER: Monaco.languages.CompletionItemKind.Value,
    L3HEADER: Monaco.languages.CompletionItemKind.Value,
    L4HEADER: Monaco.languages.CompletionItemKind.Value,
    L4SUBSECTION: Monaco.languages.CompletionItemKind.Value,
    L5HEADER: Monaco.languages.CompletionItemKind.Value,
    L5SUBSECTION: Monaco.languages.CompletionItemKind.Value,
    L6HEADER: Monaco.languages.CompletionItemKind.Value,
    LABEL: Monaco.languages.CompletionItemKind.Value,
    LANGDEF: Monaco.languages.CompletionItemKind.Value,
    LANGSTR: Monaco.languages.CompletionItemKind.Value,
    LIBRARY: Monaco.languages.CompletionItemKind.Value,
    LIST: Monaco.languages.CompletionItemKind.Value,
    LITERAL: Monaco.languages.CompletionItemKind.Value,
    LOCAL: Monaco.languages.CompletionItemKind.Value,
    LOCALVAR: Monaco.languages.CompletionItemKind.Value,
    LOCALVARIABLE: Monaco.languages.CompletionItemKind.Value,
    LOGGERSECTION: Monaco.languages.CompletionItemKind.Value,
    LTLIBRARY: Monaco.languages.CompletionItemKind.Value,
    MACRO: Monaco.languages.CompletionItemKind.Value,
    MACROFILE: Monaco.languages.CompletionItemKind.Value,
    MACROPARAM: Monaco.languages.CompletionItemKind.Value,
    MACROPARAMETER: Monaco.languages.CompletionItemKind.Value,
    MAINMENU: Monaco.languages.CompletionItemKind.Value,
    MAKEFILE: Monaco.languages.CompletionItemKind.Value,
    MAN: Monaco.languages.CompletionItemKind.Value,
    MANUAL: Monaco.languages.CompletionItemKind.Value,
    MAP: Monaco.languages.CompletionItemKind.Value,
    MASTERSTHESIS: Monaco.languages.CompletionItemKind.Value,
    MATCHEDTEMPLATE: Monaco.languages.CompletionItemKind.Value,
    MEMBER: Monaco.languages.CompletionItemKind.Value,
    MENU: Monaco.languages.CompletionItemKind.Value,
    MESSAGE: Monaco.languages.CompletionItemKind.Value,
    METHOD: Monaco.languages.CompletionItemKind.Method,
    METHODSPEC: Monaco.languages.CompletionItemKind.Value,
    MINORMODE: Monaco.languages.CompletionItemKind.Value,
    MISC: Monaco.languages.CompletionItemKind.Value,
    MIXIN: Monaco.languages.CompletionItemKind.Value,
    MLCONN: Monaco.languages.CompletionItemKind.Value,
    MLPROP: Monaco.languages.CompletionItemKind.Value,
    MLTABLE: Monaco.languages.CompletionItemKind.Value,
    MODIFIEDFILE: Monaco.languages.CompletionItemKind.Value,
    MODPORT: Monaco.languages.CompletionItemKind.Value,
    MODULE: Monaco.languages.CompletionItemKind.Module,
    MODULEPAR: Monaco.languages.CompletionItemKind.Value,
    MULTITASK: Monaco.languages.CompletionItemKind.Value,
    MXTAG: Monaco.languages.CompletionItemKind.Value,
    NAME: Monaco.languages.CompletionItemKind.Value,
    NAMEATTR: Monaco.languages.CompletionItemKind.Value,
    NAMEDPATTERN: Monaco.languages.CompletionItemKind.Value,
    NAMEDTEMPLATE: Monaco.languages.CompletionItemKind.Value,
    NAMELIST: Monaco.languages.CompletionItemKind.Value,
    NAMESPACE: Monaco.languages.CompletionItemKind.Module,
    NET: Monaco.languages.CompletionItemKind.Value,
    NETTYPE: Monaco.languages.CompletionItemKind.Value,
    NEWFILE: Monaco.languages.CompletionItemKind.Value,
    NODE: Monaco.languages.CompletionItemKind.Value,
    NOTATION: Monaco.languages.CompletionItemKind.Value,
    NSPREFIX: Monaco.languages.CompletionItemKind.Value,
    NULL: Monaco.languages.CompletionItemKind.Value,
    NUMBER: Monaco.languages.CompletionItemKind.Value,
    OBJECT: Monaco.languages.CompletionItemKind.Value,
    ONEOF: Monaco.languages.CompletionItemKind.Value,
    OPARAM: Monaco.languages.CompletionItemKind.Value,
    OPERATOR: Monaco.languages.CompletionItemKind.Operator,
    OPTENABLE: Monaco.languages.CompletionItemKind.Value,
    OPTION: Monaco.languages.CompletionItemKind.Value,
    OPTWITH: Monaco.languages.CompletionItemKind.Value,
    PACKAGE: Monaco.languages.CompletionItemKind.Module,
    PACKAGENAME: Monaco.languages.CompletionItemKind.Value,
    PACKSPEC: Monaco.languages.CompletionItemKind.Value,
    PARAGRAPH: Monaco.languages.CompletionItemKind.Value,
    PARAM: Monaco.languages.CompletionItemKind.Value,
    PARAMETER: Monaco.languages.CompletionItemKind.TypeParameter,
    PARAMETERENTITY: Monaco.languages.CompletionItemKind.Value,
    PART: Monaco.languages.CompletionItemKind.Value,
    PATCH: Monaco.languages.CompletionItemKind.Value,
    PATH: Monaco.languages.CompletionItemKind.Value,
    PATTERN: Monaco.languages.CompletionItemKind.Value,
    PHANDLER: Monaco.languages.CompletionItemKind.Value,
    PHDTHESIS: Monaco.languages.CompletionItemKind.Value,
    PKG: Monaco.languages.CompletionItemKind.Value,
    PLACEHOLDER: Monaco.languages.CompletionItemKind.Value,
    PLAY: Monaco.languages.CompletionItemKind.Value,
    PORT: Monaco.languages.CompletionItemKind.Value,
    PROBE: Monaco.languages.CompletionItemKind.Value,
    PROCEDURE: Monaco.languages.CompletionItemKind.Value,
    PROCEEDINGS: Monaco.languages.CompletionItemKind.Value,
    PROCESS: Monaco.languages.CompletionItemKind.Value,
    PROGRAM: Monaco.languages.CompletionItemKind.Value,
    PROJECT: Monaco.languages.CompletionItemKind.Value,
    PROPERTY: Monaco.languages.CompletionItemKind.Property,
    PROTECTED: Monaco.languages.CompletionItemKind.Value,
    PROTECTSPEC: Monaco.languages.CompletionItemKind.Value,
    PROTOCOL: Monaco.languages.CompletionItemKind.Value,
    PROTODEF: Monaco.languages.CompletionItemKind.Value,
    PROTOTYPE: Monaco.languages.CompletionItemKind.Value,
    PUBLICATION: Monaco.languages.CompletionItemKind.Value,
    QMP: Monaco.languages.CompletionItemKind.Value,
    QUALNAME: Monaco.languages.CompletionItemKind.Value,
    RECEIVER: Monaco.languages.CompletionItemKind.Value,
    RECORD: Monaco.languages.CompletionItemKind.Value,
    RECORDFIELD: Monaco.languages.CompletionItemKind.Value,
    REGEX: Monaco.languages.CompletionItemKind.Value,
    REGION: Monaco.languages.CompletionItemKind.Value,
    REGISTER: Monaco.languages.CompletionItemKind.Value,
    REOPEN: Monaco.languages.CompletionItemKind.Value,
    REPOID: Monaco.languages.CompletionItemKind.Value,
    REPOSITORYID: Monaco.languages.CompletionItemKind.Value,
    REPR: Monaco.languages.CompletionItemKind.Value,
    RESOURCE: Monaco.languages.CompletionItemKind.Value,
    RESPONSE: Monaco.languages.CompletionItemKind.Value,
    ROLE: Monaco.languages.CompletionItemKind.Value,
    ROOT: Monaco.languages.CompletionItemKind.Value,
    RPC: Monaco.languages.CompletionItemKind.Value,
    RULE: Monaco.languages.CompletionItemKind.Value,
    RUN: Monaco.languages.CompletionItemKind.Value,
    SCHEMA: Monaco.languages.CompletionItemKind.Value,
    SCRIPT: Monaco.languages.CompletionItemKind.Value,
    SECTION: Monaco.languages.CompletionItemKind.Value,
    SECTIONGROUP: Monaco.languages.CompletionItemKind.Value,
    SELECTOR: Monaco.languages.CompletionItemKind.Value,
    SEQUENCE: Monaco.languages.CompletionItemKind.Value,
    SERVER: Monaco.languages.CompletionItemKind.Value,
    SERVICE: Monaco.languages.CompletionItemKind.Value,
    SET: Monaco.languages.CompletionItemKind.Value,
    SETTER: Monaco.languages.CompletionItemKind.Value,
    SIGNAL: Monaco.languages.CompletionItemKind.Value,
    SIGNATURE: Monaco.languages.CompletionItemKind.Value,
    SINGLETONMETHOD: Monaco.languages.CompletionItemKind.Value,
    SLOT: Monaco.languages.CompletionItemKind.Value,
    SOURCE: Monaco.languages.CompletionItemKind.Value,
    SOURCEFILE: Monaco.languages.CompletionItemKind.Value,
    STEP: Monaco.languages.CompletionItemKind.Value,
    STRING: Monaco.languages.CompletionItemKind.Value,
    STRUCT: Monaco.languages.CompletionItemKind.Struct,
    STRUCTURE: Monaco.languages.CompletionItemKind.Value,
    STYLESHEET: Monaco.languages.CompletionItemKind.Value,
    SUBDIR: Monaco.languages.CompletionItemKind.Value,
    SUBMETHOD: Monaco.languages.CompletionItemKind.Value,
    SUBMODULE: Monaco.languages.CompletionItemKind.Value,
    SUBPARAGRAPH: Monaco.languages.CompletionItemKind.Value,
    SUBPROGRAM: Monaco.languages.CompletionItemKind.Value,
    SUBPROGSPEC: Monaco.languages.CompletionItemKind.Value,
    SUBROUTINE: Monaco.languages.CompletionItemKind.Value,
    SUBROUTINEDECLARATION: Monaco.languages.CompletionItemKind.Value,
    SUBSECTION: Monaco.languages.CompletionItemKind.Value,
    SUBSPEC: Monaco.languages.CompletionItemKind.Value,
    SUBST: Monaco.languages.CompletionItemKind.Value,
    SUBSTDEF: Monaco.languages.CompletionItemKind.Value,
    SUBSUBSECTION: Monaco.languages.CompletionItemKind.Value,
    SUBTITLE: Monaco.languages.CompletionItemKind.Value,
    SUBTYPE: Monaco.languages.CompletionItemKind.Value,
    SYMBOL: Monaco.languages.CompletionItemKind.Value,
    SYNONYM: Monaco.languages.CompletionItemKind.Value,
    TABLE: Monaco.languages.CompletionItemKind.Value,
    TAG: Monaco.languages.CompletionItemKind.Value,
    TALIAS: Monaco.languages.CompletionItemKind.Value,
    TARGET: Monaco.languages.CompletionItemKind.Value,
    TASK: Monaco.languages.CompletionItemKind.Value,
    TASKSPEC: Monaco.languages.CompletionItemKind.Value,
    TECHREPORT: Monaco.languages.CompletionItemKind.Value,
    TEMPLATE: Monaco.languages.CompletionItemKind.Value,
    TEST: Monaco.languages.CompletionItemKind.Value,
    TESTCASE: Monaco.languages.CompletionItemKind.Value,
    THEME: Monaco.languages.CompletionItemKind.Value,
    THEOREM: Monaco.languages.CompletionItemKind.Value,
    THRIFTFILE: Monaco.languages.CompletionItemKind.Value,
    THROWSPARAM: Monaco.languages.CompletionItemKind.Value,
    TIMER: Monaco.languages.CompletionItemKind.Value,
    TITLE: Monaco.languages.CompletionItemKind.Value,
    TOKEN: Monaco.languages.CompletionItemKind.Value,
    TOPLEVELVARIABLE: Monaco.languages.CompletionItemKind.Value,
    TPARAM: Monaco.languages.CompletionItemKind.Value,
    TRAIT: Monaco.languages.CompletionItemKind.Value,
    TRIGGER: Monaco.languages.CompletionItemKind.Value,
    TYPE: Monaco.languages.CompletionItemKind.Value,
    TYPEALIAS: Monaco.languages.CompletionItemKind.Value,
    TYPEDEF: Monaco.languages.CompletionItemKind.Value,
    TYPESPEC: Monaco.languages.CompletionItemKind.Value,
    UNION: Monaco.languages.CompletionItemKind.Value,
    UNIT: Monaco.languages.CompletionItemKind.Value,
    UNKNOWN: Monaco.languages.CompletionItemKind.Value,
    UNPUBLISHED: Monaco.languages.CompletionItemKind.Value,
    USERNAME: Monaco.languages.CompletionItemKind.Value,
    USING: Monaco.languages.CompletionItemKind.Value,
    VAL: Monaco.languages.CompletionItemKind.Value,
    VALUE: Monaco.languages.CompletionItemKind.Value,
    VAR: Monaco.languages.CompletionItemKind.Value,
    VARALIAS: Monaco.languages.CompletionItemKind.Value,
    VARIABLE: Monaco.languages.CompletionItemKind.Variable,
    VARSPEC: Monaco.languages.CompletionItemKind.Value,
    VECTOR: Monaco.languages.CompletionItemKind.Value,
    VERSION: Monaco.languages.CompletionItemKind.Value,
    VIEW: Monaco.languages.CompletionItemKind.Value,
    VIRTUAL: Monaco.languages.CompletionItemKind.Value,
    VRESOURCE: Monaco.languages.CompletionItemKind.Value,
    WRAPPER: Monaco.languages.CompletionItemKind.Value,
    XINPUT: Monaco.languages.CompletionItemKind.Value,
    XTASK: Monaco.languages.CompletionItemKind.Value,
}

/**
 * Maps a SearchMatch result from the server to a partial competion item for
 * Monaco.
 */
const suggestionToCompletionItems = (
    suggestion: SearchMatch,
    options: { globbing: boolean; filterValue?: string }
): PartialCompletionItem[] | PartialCompletionItem => {
    switch (suggestion.type) {
        case 'path':
            return {
                label: suggestion.path,
                kind: Monaco.languages.CompletionItemKind.File,
                insertText: regexInsertText(suggestion.path, options) + ' ',
                filterText: suggestion.path,
                detail: `${suggestion.path} - ${suggestion.repository}`,
            }
        case 'repo':
            return {
                label: suggestion.repository,
                kind: repositoryCompletionItemKind,
                insertText: repositoryInsertText(suggestion, options) + ' ',
                filterText: suggestion.repository,
            }
        case 'symbol':
            return suggestion.symbols.map(({ name, kind }) => ({
                label: name,
                kind: symbolKindToCompletionItemKind[kind],
                insertText: name + ' ',
                filterText: name,
                detail: `${startCase(kind.toLowerCase())} - ${suggestion.path} - ${suggestion.repository}`,
            }))
    }
    return []
}

/**
 * An internal Monaco command causing completion providers to be invoked,
 * and the suggestions widget to be shown.
 *
 * Useful to show the suggestions widget right after selecting a filter type
 * completion, to offer filter values completions.
 */
const TRIGGER_SUGGESTIONS: Monaco.languages.Command = {
    id: 'editor.action.triggerSuggest',
    title: 'Trigger suggestions',
}

/**
 * completeDefault returns completions for filters or symbols.
 * If the query is empty (no token is availabe) or if completion is triggered at
 * a whitespace, it returns a static list of available filters.
 * If completion is triggered at a pattern and the pattern does not match a
 * known filter, it returns symbol suggestions from the server.
 */
async function completeDefault(
    tokenAtPosition: Whitespace | Pattern | null,
    serverSuggestions: (token: Token, type: SearchMatch['type']) => Promise<PartialCompletionItem[]>,
    disablePatternSuggestions = false
): Promise<Monaco.languages.CompletionList> {
    // Default filter suggestions
    let suggestions = FILTER_TYPE_COMPLETIONS.map(
        (suggestion): Monaco.languages.CompletionItem => ({
            ...suggestion,
            range: toMonacoSingleLineRange(tokenAtPosition?.range || startRange),
            command: TRIGGER_SUGGESTIONS,
        })
    )

    if (tokenAtPosition?.type === 'pattern') {
        // If the token being typed matches a known filter, only return static
        // filter type suggestions.
        // This avoids blocking on dynamic suggestions to display the
        // suggestions widget.
        if (
            suggestions.some(
                ({ label }) => typeof label === 'string' && label.startsWith(tokenAtPosition.value.toLowerCase())
            )
        ) {
            return { suggestions }
        }

        if (!disablePatternSuggestions) {
            suggestions = suggestions.concat(
                (await serverSuggestions(tokenAtPosition, 'symbol')).map(completionItem => ({
                    ...completionItem,
                    range: toMonacoSingleLineRange(tokenAtPosition.range),
                    // Set a sortText so that dynamic suggestions
                    // are shown after filter type suggestions.
                    sortText: '1',
                    command: COMPLETION_ITEM_SELECTED,
                }))
            )
        }
    }

    return { suggestions }
}

/**
 * completeFilterValue returns suggestions for filters with static values (e.g.
 * "case") or, if the filter is configured as such, fetches suggestions from the
 * server.
 */
async function completeFilterValue(
    resolvedFilter: NonNullable<ResolvedFilter>,
    token: Filter,
    serverSuggestions: (token: Filter, type: SearchMatch['type']) => Promise<PartialCompletionItem[]>,
    column: number,
    isSourcegraphDotCom?: boolean
): Promise<Monaco.languages.CompletionItem[]> {
    const defaultRange = {
        startLineNumber: 1,
        endLineNumber: 1,
        startColumn: column,
        endColumn: column,
    }
    const { value } = token

    // FIXME(tsenart): We need to refactor completions to work with
    // complex predicates.
    // For now we just disable all static suggestions for a predicate's filter
    // if we are inside that predicate.
    const insidePredicate = value ? PREDICATE_REGEX.test(value.value) : false

    let suggestions: Monaco.languages.CompletionItem[] = []
    if (resolvedFilter.definition.discreteValues && !insidePredicate) {
        suggestions = resolvedFilter.definition.discreteValues(token.value, isSourcegraphDotCom).map(
            ({ label, insertText, asSnippet }, index): Monaco.languages.CompletionItem => ({
                label,
                sortText: index.toString().padStart(2, '1'), // suggestions sort by order in the list, not alphabetically (up to 99 values).
                kind: Monaco.languages.CompletionItemKind.Value,
                insertText: `${insertText || label} `,
                filterText: label,
                insertTextRules: asSnippet ? Monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet : undefined,
                range: value ? toMonacoSingleLineRange(value.range) : defaultRange,
                command: COMPLETION_ITEM_SELECTED,
            })
        )
    }
    if (isSourcegraphDotCom === true && (value === undefined || (value.type === 'literal' && value.value === ''))) {
        // On Sourcegraph.com, prompt only static suggestions if there is no value to use for generating dynamic suggestions yet.
        return suggestions
    }
    if (resolvedFilter.definition.suggestions) {
        // If the filter definition has an associated dynamic suggestion type,
        // use it to retrieve dynamic suggestions from the backend.
        suggestions = suggestions.concat(
            (await serverSuggestions(token, resolvedFilter.definition.suggestions)).map(
                (partialCompletionItem, index) => ({
                    ...partialCompletionItem,
                    // Set the current value as filterText, so that all dynamic suggestions
                    // returned by the server are displayed. Otherwise, if the current filter value
                    // is a regex pattern like `repo:^` Monaco's filtering will try match `^` against the
                    // suggestions, and not display them because they don't match.
                    filterText: value?.value,
                    sortText: index.toString().padStart(2, '0'), // suggestions sort by order in the list, not alphabetically (up to 99 values).
                    range: value ? toMonacoSingleLineRange(value.range) : defaultRange,
                    command: COMPLETION_ITEM_SELECTED,
                })
            )
        )
    }
    return suggestions
}

const startRange: CharacterRange = { start: 1, end: 1 }

export type FetchSuggestions = <T extends SearchMatch['type']>(
    token: Token,
    type: T
) => Promise<Extract<SearchMatch, { type: T }>[]>

/**
 * Returns the completion items for a search query being typed in the Monaco query input,
 * including both static and dynamically fetched suggestions.
 */
export async function getCompletionItems(
    tokenAtPosition: Token | null,
    { column }: Pick<Monaco.Position, 'column'>,
    fetchDynamicSuggestions: FetchSuggestions,
    {
        globbing = false,
        isSourcegraphDotCom = false,
        disablePatternSuggestions = false,
    }: { globbing?: boolean; isSourcegraphDotCom?: boolean; disablePatternSuggestions?: boolean }
): Promise<Monaco.languages.CompletionList | null> {
    let suggestions: Monaco.languages.CompletionItem[] = []

    // Show all filter suggestions if the query is empty or when the token at
    // column is labeled as a pattern or whitespace, and none of filter,
    // operator, nor quoted value, show static filter type suggestions, followed
    // by dynamic suggestions.
    if (!tokenAtPosition || tokenAtPosition.type === 'pattern' || tokenAtPosition.type === 'whitespace') {
        return completeDefault(
            tokenAtPosition,
            async (token, type) =>
                (await fetchDynamicSuggestions(token, type)).flatMap(suggestion =>
                    suggestionToCompletionItems(suggestion, { globbing })
                ),
            disablePatternSuggestions
        )
    }

    if (tokenAtPosition?.type === 'filter') {
        const completingValue = !tokenAtPosition.value || tokenAtPosition.value.range.start + 1 <= column
        if (!completingValue) {
            return null
        }

        const resolvedFilter = resolveFilter(tokenAtPosition.field.value)
        if (!resolvedFilter) {
            return null
        }

        suggestions = suggestions.concat(
            await completeFilterValue(
                resolvedFilter,
                tokenAtPosition,
                async (token, type) =>
                    (await fetchDynamicSuggestions(token, type)).flatMap(suggestion =>
                        suggestionToCompletionItems(suggestion, { globbing, filterValue: token.value?.value })
                    ),
                column,
                isSourcegraphDotCom
            )
        )
        return { suggestions }
    }

    return null
}
