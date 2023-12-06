import React, { useCallback, useMemo } from 'react'

import { Prec } from '@codemirror/state'
import { keymap } from '@codemirror/view'
import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom'
import { BehaviorSubject, of } from 'rxjs'
import { debounceTime } from 'rxjs/operators'

import {
    StreamingSearchResultsList,
    type StreamingSearchResultsListProps,
    CodeMirrorQueryInput,
    createDefaultSuggestions,
    changeListener,
} from '@sourcegraph/branded'
import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { LoadingSpinner, Button, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { SearchPatternType } from '../graphql-operations'
import type { OwnConfigProps } from '../own/OwnConfigProps'
import { setSearchMode, useNavbarQueryState } from '../stores'

import { parseSearchURLQuery, parseSearchURLPatternType, type SearchStreamingProps } from '.'
import { submitSearch } from './helpers'

import styles from './SearchConsolePage.module.scss'

interface SearchConsolePageProps
    extends SearchStreamingProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded' | 'executedQuery' | 'showSearchContext'>,
        OwnConfigProps,
        TelemetryV2Props {
    isMacPlatform: boolean
}

export const SearchConsolePage: React.FunctionComponent<React.PropsWithChildren<SearchConsolePageProps>> = props => {
    const location = useLocation()
    const navigate = useNavigate()
    const { streamSearch, isSourcegraphDotCom } = props
    const { applySuggestionsOnEnter } = useExperimentalFeatures(features => ({
        applySuggestionsOnEnter: features.applySearchQuerySuggestionOnEnter ?? true,
    }))
    const searchQuery = useMemo(
        () => new BehaviorSubject<string>(parseSearchURLQuery(location.search) ?? ''),
        [location.search]
    )

    const patternType = useMemo(
        () => parseSearchURLPatternType(location.search) || SearchPatternType.structural,
        [location.search]
    )

    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const searchMode = useNavbarQueryState(state => state.searchMode)
    const submittedURLQuery = useNavbarQueryState(state => state.searchQueryFromURL)

    const triggerSearch = useCallback(() => {
        navigate('/search/console?q=' + encodeURIComponent(searchQuery.value))
    }, [navigate, searchQuery])

    const transformedQuery = useMemo(() => {
        let query = parseSearchURLQuery(location.search)
        query = query?.replace(/\/\/.*/g, '') || ''

        return query
    }, [location.search])

    const autocompletion = useMemo(
        () =>
            createDefaultSuggestions({
                fetchSuggestions: query => fetchStreamSuggestions(query),
                isSourcegraphDotCom,
                applyOnEnter: applySuggestionsOnEnter,
            }),
        [isSourcegraphDotCom, applySuggestionsOnEnter]
    )

    const extensions = useMemo(
        () => [
            Prec.highest(keymap.of([{ key: 'Mod-Enter', run: () => (triggerSearch(), true) }])),
            changeListener(value => searchQuery.next(value)),
            autocompletion,
        ],
        [searchQuery, triggerSearch, autocompletion]
    )

    // Fetch search results when the `q` URL query parameter changes
    const results = useObservable(
        useMemo(
            () =>
                streamSearch(of(transformedQuery), {
                    version: LATEST_VERSION,
                    patternType: patternType ?? SearchPatternType.standard,
                    caseSensitive: false,
                    trace: undefined,
                }).pipe(debounceTime(500)),
            [patternType, transformedQuery, streamSearch]
        )
    )

    return (
        <div className="w-100 p-2">
            <PageTitle title="Search console" />
            <div className="d-flex overflow-hidden h-100">
                <div className="flex-1 p-1 d-flex flex-column">
                    <div className={styles.editor}>
                        <CodeMirrorQueryInput
                            className="d-flex flex-column overflow-hidden"
                            patternType={patternType}
                            interpretComments={true}
                            value={searchQuery.value}
                            extensions={extensions}
                        />
                    </div>
                    <Button className="mt-2" onClick={triggerSearch} variant="primary">
                        Search &nbsp; {props.isMacPlatform ? <kbd>⌘</kbd> : <kbd>Ctrl</kbd>}+<kbd>⏎</kbd>
                    </Button>
                </div>
                <div className={classNames('flex-1 p-1', styles.results)}>
                    {results &&
                        (results.state === 'loading' ? (
                            <LoadingSpinner />
                        ) : (
                            <StreamingSearchResultsList
                                {...props}
                                allExpanded={false}
                                results={results}
                                executedQuery={location.search}
                                searchMode={searchMode}
                                setSearchMode={setSearchMode}
                                submitSearch={submitSearch}
                                caseSensitive={caseSensitive}
                                searchQueryFromURL={submittedURLQuery}
                            />
                        ))}
                </div>
            </div>
        </div>
    )
}
