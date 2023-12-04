import React, { useCallback, useMemo } from 'react'

import { EditorView } from '@codemirror/view'

import { createDefaultSuggestions, RepoFileLink } from '@sourcegraph/branded'
import { getFileMatchUrl, getRepositoryUrl, type SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { Button, Code } from '@sourcegraph/wildcard'

import type { BlockProps, SymbolBlockInput } from '../..'
import { SearchTypeSuggestionsInput } from '../suggestions/SearchTypeSuggestionsInput'
import { fetchSuggestions } from '../suggestions/suggestions'

import styles from './NotebookSymbolBlockInput.module.scss'

interface NotebookSymbolBlockInputProps extends Pick<BlockProps, 'onRunBlock'> {
    id: string
    queryInput: string
    onEditorCreated: (editor: EditorView) => void
    setQueryInput: (value: string) => void
    onSymbolSelected: (symbol: SymbolBlockInput) => void
    isSourcegraphDotCom: boolean
}

function getSymbolSuggestionsQuery(queryInput: string): string {
    return `${queryInput} fork:yes type:symbol count:50`
}

const editorAttributes = [
    EditorView.editorAttributes.of({
        'data-testid': 'notebook-symbol-block-input',
    }),
    EditorView.contentAttributes.of({
        'arial-label': 'Symbol search input',
    }),
]

export const NotebookSymbolBlockInput: React.FunctionComponent<
    React.PropsWithChildren<NotebookSymbolBlockInputProps>
> = ({ onSymbolSelected, isSourcegraphDotCom, ...inputProps }) => {
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

    const queryCompletion = useMemo(
        () =>
            createDefaultSuggestions({
                isSourcegraphDotCom,
                fetchSuggestions: fetchStreamSuggestions,
                disableSymbolCompletion: true,
            }),
        [isSourcegraphDotCom]
    )

    return (
        <div className={styles.input}>
            <SearchTypeSuggestionsInput<SymbolMatch>
                label="Find a symbol using a Sourcegraph search query"
                queryPrefix="type:symbol"
                fetchSuggestions={fetchSymbolSuggestions}
                countSuggestions={countSuggestions}
                renderSuggestions={renderSuggestions}
                extension={useMemo(() => [queryCompletion, editorAttributes], [queryCompletion])}
                {...inputProps}
            />
        </div>
    )
}

const SymbolSuggestions: React.FunctionComponent<
    React.PropsWithChildren<{
        suggestions: SymbolMatch[]
        onSymbolSelected: (symbol: SymbolBlockInput) => void
    }>
> = ({ suggestions, onSymbolSelected }) => {
    const symbolKindTags = useExperimentalFeatures(features => features.symbolKindTags)
    return (
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
                            <SymbolKind kind={symbol.kind} className="mr-1" symbolKindTags={symbolKindTags} />
                            <Code>
                                {symbol.name}{' '}
                                {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                            </Code>
                        </Button>
                    ))}
                </div>
            ))}
        </div>
    )
}
