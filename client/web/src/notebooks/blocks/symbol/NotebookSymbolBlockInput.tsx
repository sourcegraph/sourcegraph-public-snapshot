import React, { useCallback, useEffect, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import * as Monaco from 'monaco-editor'
import { of } from 'rxjs'
import { delay, startWith } from 'rxjs/operators'

import { pluralize } from '@sourcegraph/common'
import { createQueryExampleFromString, updateQueryWithFilterAndExample } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { toMonacoSelection } from '@sourcegraph/search-ui/src/input/MonacoQueryInput'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { toMonacoRange } from '@sourcegraph/shared/src/search/query/monaco'
import { getFileMatchUrl, getRepositoryUrl, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useObservable } from '@sourcegraph/wildcard'

import { BlockProps, SymbolBlockInput } from '../..'
import { fetchSuggestions } from '../suggestions'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from '../useMonacoBlockInput'

import blockStyles from '../NotebookBlock.module.scss'
import styles from './NotebookSymbolBlockInput.module.scss'

interface NotebookSymbolBlockInputProps extends ThemeProps, Pick<BlockProps, 'onRunBlock' | 'onSelectBlock'> {
    id: string
    sourcegraphSearchLanguageId: string
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    symbolQueryInput: string
    setEditor: (editor: Monaco.editor.IStandaloneCodeEditor) => void
    setSymbolQueryInput: (value: string) => void
    debouncedSetSymbolQueryInput: (value: string) => void
    onSymbolSelected: (symbol: SymbolBlockInput) => void
    setIsInputFocused: (isFocused: boolean) => void
}

function getSymbolSuggestionsQuery(queryInput: string): string {
    return `${queryInput} fork:yes type:symbol count:50`
}

const LOADING = 'LOADING' as const
const QUERY_EXAMPLE = createQueryExampleFromString('{enter-regexp-pattern}')

export const NotebookSymbolBlockInput: React.FunctionComponent<NotebookSymbolBlockInputProps> = ({
    sourcegraphSearchLanguageId,
    id,
    symbolQueryInput,
    isLightTheme,
    editor,
    setEditor,
    setSymbolQueryInput,
    debouncedSetSymbolQueryInput,
    onSymbolSelected,
    setIsInputFocused,
    ...props
}) => {
    const { isInputFocused } = useMonacoBlockInput({
        editor,
        id,
        onInputChange: debouncedSetSymbolQueryInput,
        preventNewLine: true,
        ...props,
    })

    useEffect(() => {
        // setTimeout executes the editor focus in a separate run-loop which prevents adding a newline at the start of the input,
        // if Enter key was used to show the inputs.
        setTimeout(() => editor?.focus(), 0)
    }, [editor])

    useEffect(() => {
        setIsInputFocused(isInputFocused)
        return () => setIsInputFocused(false)
    }, [isInputFocused, setIsInputFocused])

    const addExampleFilter = useCallback(
        (filterType: FilterType) => {
            const { query, placeholderRange, filterRange } = updateQueryWithFilterAndExample(
                symbolQueryInput,
                filterType,
                QUERY_EXAMPLE,
                { singular: false, negate: false, emptyValue: false }
            )
            setSymbolQueryInput(query)
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
        [editor, symbolQueryInput, setSymbolQueryInput]
    )

    const symbolSuggestions = useObservable(
        useMemo(
            () =>
                symbolQueryInput.length > 0
                    ? fetchSuggestions(
                          getSymbolSuggestionsQuery(symbolQueryInput),
                          (suggestion): suggestion is SymbolMatch => suggestion.type === 'symbol',
                          symbol => symbol
                      ).pipe(
                          // A small delay to prevent flickering loading spinner.
                          delay(300),
                          startWith(LOADING)
                      )
                    : of(undefined),
            [symbolQueryInput]
        )
    )

    const symbolSuggestionsCount = useMemo(() => {
        if (!symbolSuggestions || symbolSuggestions === LOADING) {
            return undefined
        }
        return symbolSuggestions.reduce((count, suggestion) => count + suggestion.symbols.length, 0)
    }, [symbolSuggestions])

    return (
        <div className={styles.input}>
            <div>
                <label htmlFor={`${id}-symbol-query-input`}>Find a symbol using a Sourcegraph search query</label>
                <div
                    id={`${id}-symbol-query-input`}
                    className={classNames(
                        blockStyles.monacoWrapper,
                        isInputFocused && blockStyles.selected,
                        styles.queryInputMonacoWrapper
                    )}
                >
                    <div className="d-flex">
                        <SyntaxHighlightedSearchQuery className={styles.typeSymbolQueryPart} query="type:symbol" />
                    </div>
                    <div className="flex-1">
                        <MonacoEditor
                            language={sourcegraphSearchLanguageId}
                            value={symbolQueryInput}
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
                        Filter symbols by repository
                    </Button>
                    <Button variant="secondary" size="sm" onClick={() => addExampleFilter(FilterType.file)}>
                        Filter symbols by file path
                    </Button>
                </div>
                <div className="mt-3 mb-1">
                    {symbolSuggestionsCount !== undefined && (
                        <strong>
                            {symbolSuggestionsCount} matching {pluralize('symbol', symbolSuggestionsCount)} found
                        </strong>
                    )}
                    {symbolSuggestions === LOADING && <strong>Searching for symbols...</strong>}
                </div>
                {symbolSuggestions && symbolSuggestions !== LOADING && (
                    <SymbolSuggestions suggestions={symbolSuggestions} onSymbolSelected={onSymbolSelected} />
                )}
            </div>
        </div>
    )
}

const SymbolSuggestions: React.FunctionComponent<{
    suggestions: SymbolMatch[]
    onSymbolSelected: (symbol: SymbolBlockInput) => void
}> = ({ suggestions, onSymbolSelected }) => (
    <div className={styles.symbolSuggestions}>
        {suggestions.map(suggestion => (
            <div key={`${suggestion.repository}_${suggestion.path}`} className="pr-2">
                <RepoFileLink
                    className="my-2"
                    repoName={suggestion.repository}
                    repoURL={getRepositoryUrl(suggestion.repository, suggestion.branches)}
                    filePath={suggestion.path}
                    fileURL={getFileMatchUrl(suggestion)}
                />
                {suggestion.symbols.map((symbol, index) => (
                    <Button
                        className={styles.symbolButton}
                        // We have to use the index as key in case of duplicate symbols.
                        // eslint-disable-next-line react/no-array-index-key
                        key={`${suggestion.repository}_${suggestion.path}_${symbol.containerName}_${symbol.name}_${index}`}
                        onClick={() =>
                            onSymbolSelected({
                                repositoryName: suggestion.repository,
                                filePath: suggestion.path,
                                revision: suggestion.commit ?? '',
                                symbolName: symbol.name,
                                symbolKind: symbol.kind,
                                symbolContainerName: symbol.containerName,
                                lineContext: 3,
                            })
                        }
                        data-testid="symbol-suggestion-button"
                    >
                        <SymbolIcon kind={symbol.kind} className="icon-inline mr-1" />
                        <code>
                            {symbol.name}{' '}
                            {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                        </code>
                    </Button>
                ))}
            </div>
        ))}
    </div>
)
