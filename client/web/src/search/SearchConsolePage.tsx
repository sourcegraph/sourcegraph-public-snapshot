import React, { useCallback, useMemo } from 'react'

import { Prec } from '@codemirror/state'
import { keymap } from '@codemirror/view'
import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom-v5-compat'
import { BehaviorSubject } from 'rxjs'
import { debounceTime } from 'rxjs/operators'

import {
    StreamingSearchResultsList,
    StreamingSearchResultsListProps,
    CodeMirrorQueryInput,
    createDefaultSuggestions,
    changeListener,
} from '@sourcegraph/branded'
import { transformSearchQuery } from '@sourcegraph/shared/src/api/client/search'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { LoadingSpinner, Button, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { SearchPatternType } from '../graphql-operations'
import { useExperimentalFeatures } from '../stores'
import { eventLogger } from '../tracking/eventLogger'

import { parseSearchURLQuery, parseSearchURLPatternType, SearchStreamingProps } from '.'

import styles from './SearchConsolePage.module.scss'

interface SearchConsolePageProps
    extends SearchStreamingProps,
        Omit<
            StreamingSearchResultsListProps,
            'allExpanded' | 'extensionsController' | 'executedQuery' | 'showSearchContext'
        >,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI'> {
    globbing: boolean
    isMacPlatform: boolean
}

export const SearchConsolePage: React.FunctionComponent<React.PropsWithChildren<SearchConsolePageProps>> = props => {
    const location = useLocation()
    const navigate = useNavigate()
    const { globbing, streamSearch, extensionsController, isSourcegraphDotCom } = props
    const extensionHostAPI =
        extensionsController !== null && window.context.enableLegacyExtensions ? extensionsController.extHostAPI : null
    const enableGoImportsSearchQueryTransform = useExperimentalFeatures(
        features => features.enableGoImportsSearchQueryTransform
    )
    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true

    const searchQuery = useMemo(
        () => new BehaviorSubject<string>(parseSearchURLQuery(location.search) ?? ''),
        [location.search]
    )

    const patternType = useMemo(
        () => parseSearchURLPatternType(location.search) || SearchPatternType.structural,
        [location.search]
    )

    const triggerSearch = useCallback(() => {
        navigate('/search/console?q=' + encodeURIComponent(searchQuery.value))
    }, [navigate, searchQuery])

    const transformedQuery = useMemo(() => {
        let query = parseSearchURLQuery(location.search)
        query = query?.replace(/\/\/.*/g, '') || ''

        return transformSearchQuery({
            query,
            extensionHostAPIPromise: extensionHostAPI,
            enableGoImportsSearchQueryTransform,
            eventLogger,
        })
    }, [location.search, extensionHostAPI, enableGoImportsSearchQueryTransform])

    const autocompletion = useMemo(
        () =>
            createDefaultSuggestions({
                fetchSuggestions: query => fetchStreamSuggestions(query),
                globbing,
                isSourcegraphDotCom,
                applyOnEnter: applySuggestionsOnEnter,
            }),
        [globbing, isSourcegraphDotCom, applySuggestionsOnEnter]
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
                streamSearch(transformedQuery, {
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
                            isLightTheme={props.isLightTheme}
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
                                assetsRoot={window.context?.assetsRoot || ''}
                                executedQuery={location.search}
                            />
                        ))}
                </div>
            </div>
        </div>
    )
}
