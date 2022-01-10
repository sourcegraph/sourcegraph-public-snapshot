import classNames from 'classnames'
import PencilOutlineIcon from 'mdi-react/PencilOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ISavedSearch } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { AuthenticatedUser } from '../../auth'
import { buildSearchURLQueryFromQueryState } from '../../stores'

import { ActionButtonGroup } from './ActionButtonGroup'
import { EmptyPanelContainer } from './EmptyPanelContainer'
import { FooterPanel } from './FooterPanel'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchSavedSearches: () => Observable<ISavedSearch[]>
}

export const SavedSearchesPanel: React.FunctionComponent<Props> = ({
    authenticatedUser,
    fetchSavedSearches,
    className,
    telemetryService,
}) => {
    const savedSearches = useObservable(useMemo(() => fetchSavedSearches(), [fetchSavedSearches]))
    const [showAllSearches, setShowAllSearches] = useState(true)

    useEffect(() => {
        // Only log the first load (when items to load is equal to the page size)
        if (savedSearches) {
            telemetryService.log(
                'SavedSearchesPanelLoaded',
                { empty: savedSearches.length === 0, showAllSearches },
                { empty: savedSearches.length === 0, showAllSearches }
            )
        }
    }, [savedSearches, telemetryService, showAllSearches])

    const logEvent = useCallback((event: string, props?: any) => (): void => telemetryService.log(event, props), [
        telemetryService,
    ])

    const emptyDisplay = (
        <EmptyPanelContainer className="text-muted">
            <small>
                Use saved searches to alert you to uses of a favorite API, or changes to code you need to monitor.
            </small>
            {authenticatedUser && (
                <Link
                    to={`/users/${authenticatedUser.username}/searches/add`}
                    onClick={logEvent('SavedSearchesPanelCreateButtonClicked', { source: 'empty view' })}
                    className="btn btn-secondary mt-2 align-self-center"
                >
                    <PlusIcon className="icon-inline" />
                    Create a saved search
                </Link>
            )}
        </EmptyPanelContainer>
    )
    const loadingDisplay = <LoadingPanelView text="Loading saved searches" />

    const contentDisplay = (
        <div className="d-flex flex-column h-100">
            <div className="d-flex justify-content-between mb-1 mt-2">
                <small>Search</small>
                <small>Edit</small>
            </div>
            <dl className="list-group-flush flex-grow-1">
                {savedSearches
                    ?.filter(search => (showAllSearches ? true : search.namespace.id === authenticatedUser?.id))
                    .map(search => (
                        <dd key={search.id} className="text-monospace test-saved-search-entry">
                            <div className="d-flex justify-content-between">
                                <small>
                                    <Link
                                        to={'/search?' + buildSearchURLQueryFromQueryState({ query: search.query })}
                                        className=" p-0"
                                        onClick={logEvent('SavedSearchesPanelSearchClicked')}
                                    >
                                        {search.description}
                                    </Link>
                                </small>
                                {authenticatedUser &&
                                    (search.namespace.__typename === 'User' ? (
                                        <Link
                                            to={`/users/${search.namespace.namespaceName}/searches/${search.id}`}
                                            onClick={logEvent('SavedSearchesPanelEditClicked')}
                                        >
                                            <PencilOutlineIcon className="icon-inline" />
                                        </Link>
                                    ) : (
                                        <Link
                                            to={`/organizations/${search.namespace.namespaceName}/searches/${search.id}`}
                                            onClick={logEvent('SavedSearchesPanelEditClicked')}
                                        >
                                            <PencilOutlineIcon className="icon-inline" />
                                        </Link>
                                    ))}
                            </div>
                        </dd>
                    ))}
            </dl>
            {authenticatedUser && (
                <FooterPanel className="p-1">
                    <small>
                        <Link
                            to={`/users/${authenticatedUser.username}/searches`}
                            className=" text-left"
                            onClick={logEvent('SavedSearchesPanelViewAllClicked')}
                        >
                            View saved searches
                        </Link>
                    </small>
                </FooterPanel>
            )}
        </div>
    )

    const actionButtons = (
        <ActionButtonGroup>
            <div className="btn-group btn-group-sm">
                {authenticatedUser && (
                    <Link
                        to={`/users/${authenticatedUser.username}/searches/add`}
                        className="btn btn-outline-secondary mr-2"
                        onClick={logEvent('SavedSearchesPanelCreateButtonClicked', { source: 'toolbar' })}
                    >
                        +
                    </Link>
                )}
            </div>
            <div className="btn-group btn-group-sm">
                <button
                    type="button"
                    onClick={() => setShowAllSearches(false)}
                    className={classNames('btn btn-outline-secondary test-saved-search-panel-my-searches', {
                        active: !showAllSearches,
                    })}
                >
                    My searches
                </button>
                <button
                    type="button"
                    onClick={() => setShowAllSearches(true)}
                    className={classNames('btn btn-outline-secondary test-saved-search-panel-all-searches', {
                        active: showAllSearches,
                    })}
                >
                    All searches
                </button>
            </div>
        </ActionButtonGroup>
    )
    return (
        <PanelContainer
            className={classNames(className, 'saved-searches-panel')}
            title="Saved searches"
            state={savedSearches ? (savedSearches.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
