import React, { useCallback } from 'react'

import * as Monaco from 'monaco-editor'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { getFileMatchUrl, getRepositoryUrl, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button } from '@sourcegraph/wildcard'

import { BlockProps, SymbolBlockInput } from '../..'
import { SearchTypeSuggestionsInput } from '../suggestions/SearchTypeSuggestionsInput'
import { fetchSuggestions } from '../suggestions/suggestions'
import { useFocusMonacoEditorOnMount } from '../useFocusMonacoEditorOnMount'

import styles from './NotebookSymbolBlockInput.module.scss'

interface NotebookSymbolBlockInputProps extends ThemeProps, Pick<BlockProps, 'onRunBlock'> {
    id: string
    sourcegraphSearchLanguageId: string
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    queryInput: string
    setEditor: (editor: Monaco.editor.IStandaloneCodeEditor) => void
    setQueryInput: (value: string) => void
    debouncedSetQueryInput: (value: string) => void
    onSymbolSelected: (symbol: SymbolBlockInput) => void
}

function getSymbolSuggestionsQuery(queryInput: string): string {
    return `${queryInput} fork:yes type:symbol count:50`
}

export const NotebookSymbolBlockInput: React.FunctionComponent<
    React.PropsWithChildren<NotebookSymbolBlockInputProps>
> = ({ editor, onSymbolSelected, ...props }) => {
    useFocusMonacoEditorOnMount({ editor, isEditing: true })

    const fetchSymbolSuggestions = useCallback(
        (query: string) =>
            fetchSuggestions(
                getSymbolSuggestionsQuery(query),
                (suggestion): suggestion is SymbolMatch => suggestion.type === 'symbol',
                symbol => symbol
            ),
        []
    )

    const countSuggestions = useCallback(
        (suggestions: SymbolMatch[]) => suggestions.reduce((count, suggestion) => count + suggestion.symbols.length, 0),
        []
    )

    const renderSuggestions = useCallback(
        (suggestions: SymbolMatch[]) => (
            <SymbolSuggestions suggestions={suggestions} onSymbolSelected={onSymbolSelected} />
        ),
        [onSymbolSelected]
    )

    return (
        <div className={styles.input}>
            <SearchTypeSuggestionsInput<SymbolMatch>
                label="Find a symbol using a Sourcegraph search query"
                queryPrefix="type:symbol"
                editor={editor}
                fetchSuggestions={fetchSymbolSuggestions}
                countSuggestions={countSuggestions}
                renderSuggestions={renderSuggestions}
                {...props}
            />
        </div>
    )
}

const SymbolSuggestions: React.FunctionComponent<
    React.PropsWithChildren<{
        suggestions: SymbolMatch[]
        onSymbolSelected: (symbol: SymbolBlockInput) => void
    }>
> = ({ suggestions, onSymbolSelected }) => (
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
                        <SymbolIcon kind={symbol.kind} className="mr-1" />
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
