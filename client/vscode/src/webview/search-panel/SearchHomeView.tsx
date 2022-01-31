import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'

import {
    SearchPatternType,
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { LATEST_VERSION, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'

import { SearchHomeState } from '../../state'
import { WebviewPageProps } from '../platform/context'

import { BrandHeader } from './components/BrandHeader'
import { HomeFooter } from './components/HomeFooter'
import styles from './index.module.scss'

export interface SearchHomeViewProps extends WebviewPageProps {
    context: SearchHomeState['context']
}

export const SearchHomeView: React.FunctionComponent<SearchHomeViewProps> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
    context,
    instanceURL,
}) => {
    // Toggling case sensitivity or pattern type does NOT trigger a new search on home view.
    const [caseSensitive, setCaseSensitivity] = useState(false)
    const [patternType, setPatternType] = useState(SearchPatternType.literal)

    const [userQueryState, setUserQueryState] = useState({
        query: '',
    })

    useEffect(() => {
        console.log('initial mount')
    }, [])

    // TODO: we need an API for updating search query state from the sidebar.

    const onSubmit = useCallback(() => {
        extensionCoreAPI
            .streamSearch(userQueryState.query, {
                caseSensitive,
                patternType,
                version: LATEST_VERSION,
                trace: undefined,
            })
            .then(() => {})
            .catch(() => {})
    }, [userQueryState.query, caseSensitive, patternType, extensionCoreAPI])

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string) => {
            extensionCoreAPI.setSelectedSearchContextSpec(spec).catch(error => {
                console.error('Error persisting search context spec.', error)
            })
        },
        [extensionCoreAPI]
    )

    const fetchStreamSuggestions = useCallback(
        (query): Observable<SearchMatch[]> =>
            wrapRemoteObservable(extensionCoreAPI.fetchStreamSuggestions(query, instanceURL)),
        [extensionCoreAPI, instanceURL]
    )

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    return (
        <div className="d-flex flex-column align-items-center">
            <BrandHeader theme={theme} extensionCoreAPI={extensionCoreAPI} />

            <div className={styles.homeSearchBoxContainer}>
                {/* eslint-disable-next-line react/forbid-elements */}
                <form
                    className="d-flex my-2"
                    onSubmit={event => {
                        event.preventDefault()
                        onSubmit()
                    }}
                >
                    <SearchBox
                        caseSensitive={caseSensitive}
                        setCaseSensitivity={setCaseSensitivity}
                        patternType={patternType}
                        setPatternType={setPatternType}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        hasUserAddedExternalServices={false}
                        hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                        structuralSearchDisabled={false}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        authenticatedUser={authenticatedUser}
                        searchContextsEnabled={true}
                        showSearchContext={true}
                        showSearchContextManagement={true}
                        defaultSearchContextSpec="global"
                        setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                        selectedSearchContextSpec={context.selectedSearchContextSpec}
                        fetchSearchContexts={fetchSearchContexts}
                        fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                        getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                        fetchStreamSuggestions={fetchStreamSuggestions}
                        settingsCascade={settingsCascade}
                        globbing={globbing}
                        isLightTheme={theme === 'theme-light'}
                        telemetryService={platformContext.telemetryService}
                        platformContext={platformContext}
                        className={classNames('flex-grow-1 flex-shrink-past-contents', styles.searchBox)}
                        containerClassName={styles.searchBoxContainer}
                    />
                </form>

                <HomeFooter
                    isLightTheme={theme === 'theme-light'}
                    setQuery={setUserQueryState}
                    telemetryService={platformContext.telemetryService}
                />
            </div>
        </div>
    )
}
