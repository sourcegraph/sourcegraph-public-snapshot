import { ReactElement, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import * as Monaco from 'monaco-editor'
import { Observable, of } from 'rxjs'
import { delay, startWith } from 'rxjs/operators'

import { pluralize } from '@sourcegraph/common'
import { createQueryExampleFromString, updateQueryWithFilterAndExample } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery, toMonacoSelection } from '@sourcegraph/search-ui'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { toMonacoRange } from '@sourcegraph/shared/src/search/query/monaco'
import { PathMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useObservable } from '@sourcegraph/wildcard'

import { BlockProps } from '../..'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from '../useMonacoBlockInput'

import blockStyles from '../NotebookBlock.module.scss'
import styles from './SearchTypeSuggestionsInput.module.scss'

interface SearchTypeSuggestionsInputProps<S extends SymbolMatch | PathMatch>
    extends ThemeProps,
        Pick<BlockProps, 'onRunBlock'> {
    id: string
    label: string
    sourcegraphSearchLanguageId: string
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    queryPrefix: string
    queryInput: string
    setEditor: (editor: Monaco.editor.IStandaloneCodeEditor) => void
    setQueryInput: (value: string) => void
    debouncedSetQueryInput: (value: string) => void
    fetchSuggestions: (query: string) => Observable<S[]>
    countSuggestions: (suggestions: S[]) => number
    renderSuggestions: (suggestions: S[]) => ReactElement
}

const LOADING = 'LOADING' as const
const QUERY_EXAMPLE = createQueryExampleFromString('{enter-regexp-pattern}')

export const SearchTypeSuggestionsInput = <S extends SymbolMatch | PathMatch>({
    id,
    label,
    editor,
    sourcegraphSearchLanguageId,
    queryPrefix,
    queryInput,
    isLightTheme,
    setEditor,
    setQueryInput,
    debouncedSetQueryInput,
    fetchSuggestions,
    countSuggestions,
    renderSuggestions,
    ...props
}: SearchTypeSuggestionsInputProps<S>): ReactElement => {
    useMonacoBlockInput({
        editor,
        id,
        onInputChange: debouncedSetQueryInput,
        preventNewLine: true,
        ...props,
    })

    const addExampleFilter = useCallback(
        (filterType: FilterType) => {
            const { query, placeholderRange, filterRange } = updateQueryWithFilterAndExample(
                queryInput,
                filterType,
                QUERY_EXAMPLE,
                { singular: false, negate: false, emptyValue: false }
            )
            setQueryInput(query)
            const textModel = editor?.getModel()
            if (!editor || !textModel) {
                return
            }
            // Focus the selection in the next run-loop, since we have to wait for the Monaco editor to update.
            setTimeout(() => {
                const selectionRange = toMonacoSelection(toMonacoRange(placeholderRange, textModel))
                editor.setSelection(selectionRange)
                editor.revealRange(toMonacoRange(filterRange, textModel))
                editor.focus()
            }, 0)
        },
        [editor, queryInput, setQueryInput]
    )

    const suggestions = useObservable(
        useMemo(
            () =>
                queryInput.length > 0
                    ? fetchSuggestions(queryInput).pipe(
                          // A small delay to prevent flickering loading message.
                          delay(300),
                          startWith(LOADING)
                      )
                    : of(undefined),
            [queryInput, fetchSuggestions]
        )
    )

    const suggestionsCount = useMemo(() => {
        if (!suggestions || suggestions === LOADING) {
            return undefined
        }
        return countSuggestions(suggestions)
    }, [suggestions, countSuggestions])

    return (
        <div>
            <label htmlFor={`${id}-search-type-query-input`}>{label}</label>
            <div
                id={`${id}-search-type-query-input`}
                className={classNames(blockStyles.monacoWrapper, styles.queryInputMonacoWrapper)}
            >
                <div className="d-flex">
                    <SyntaxHighlightedSearchQuery className={styles.searchTypeQueryPart} query={queryPrefix} />
                </div>
                <div className="flex-1">
                    <MonacoEditor
                        language={sourcegraphSearchLanguageId}
                        value={queryInput}
                        height={17}
                        isLightTheme={isLightTheme}
                        editorWillMount={noop}
                        onEditorCreated={setEditor}
                        options={{
                            ...MONACO_BLOCK_INPUT_OPTIONS,
                            wordWrap: 'off',
                            scrollbar: {
                                vertical: 'hidden',
                                horizontal: 'hidden',
                            },
                        }}
                        border={false}
                    />
                </div>
            </div>
            <div className="mt-1">
                <Button
                    className="mr-1"
                    variant="secondary"
                    size="sm"
                    onClick={() => addExampleFilter(FilterType.repo)}
                >
                    Filter by repository
                </Button>
                <Button variant="secondary" size="sm" onClick={() => addExampleFilter(FilterType.file)}>
                    Filter by file path
                </Button>
            </div>
            <div className="mt-3 mb-1">
                {suggestionsCount !== undefined && (
                    <strong>
                        {suggestionsCount} {pluralize('result', suggestionsCount)} found
                    </strong>
                )}
                {suggestions === LOADING && <strong>Searching...</strong>}
            </div>
            {suggestions && suggestions !== LOADING && renderSuggestions(suggestions)}
        </div>
    )
}
