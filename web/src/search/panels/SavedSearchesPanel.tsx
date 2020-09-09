import React, { useMemo, useState } from 'react'
import classNames from 'classnames'
import { PanelContainer } from './PanelContainer'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { Link } from '../../../../shared/src/components/Link'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { SearchPatternType, ISavedSearch } from '../../../../shared/src/graphql/schema'
import { AuthenticatedUser } from '../../auth'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import PencilOutlineIcon from 'mdi-react/PencilOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import { Observable } from 'rxjs'

export const SavedSearchesPanel: React.FunctionComponent<{
    patternType: SearchPatternType
    authenticatedUser: AuthenticatedUser | null
    fetchSavedSearches: () => Observable<ISavedSearch[]>
    className?: string
    /** For testing only */
    displayState?: 'loading' | 'empty' | 'populated'
    /** For testing only */
    mySearchesMode?: boolean
}> = ({ patternType, authenticatedUser, fetchSavedSearches, className, displayState, mySearchesMode }) => {
    const savedSearches = useObservable(useMemo(() => fetchSavedSearches(), [fetchSavedSearches]))
    const [showAllSearches, setShowAllSearches] = useState(true)

    const emptyDisplay = (
        <div className="panel-container__empty-container">
            <small>
                Use saved searches to alert you to uses of a favorite api, or changes to code you need to monitor.
            </small>
            {authenticatedUser && (
                <Link to={`/users/${authenticatedUser.username}/searches/add`} className="btn btn-secondary mt-2">
                    <PlusIcon className="icon-inline" />
                    Create a saved search
                </Link>
            )}
        </div>
    )
    const loadingDisplay = (
        <div className="d-flex justify-content-center align-items-center panel-container__empty-container">
            <div className="icon-inline">
                <LoadingSpinner />
            </div>
            Loading saved searches
        </div>
    )

    const contentDisplay = (
        <>
            <div>
                <div className="d-flex justify-content-between mb-1">
                    <small>Search</small>
                    <small>Edit</small>
                </div>
                <dl className="list-group-flush">
                    {savedSearches
                        ?.filter(search =>
                            showAllSearches && !mySearchesMode ? true : search.namespace.id === authenticatedUser?.id
                        )
                        .map(search => (
                            <dd key={search.id} className="text-monospace test-saved-search-entery">
                                <div className="d-flex justify-content-between">
                                    <Link
                                        to={'/search?' + buildSearchURLQuery(search.query, patternType, false)}
                                        className="btn btn-link p-0"
                                    >
                                        {search.description}
                                    </Link>
                                    {authenticatedUser &&
                                        (search.namespace.__typename === 'User' ? (
                                            <Link to={`/users/${search.namespace.namespaceName}/searches/${search.id}`}>
                                                <PencilOutlineIcon className="icon-inline" />
                                            </Link>
                                        ) : (
                                            <Link
                                                to={`/organizations/${search.namespace.namespaceName}/searches/${search.id}`}
                                            >
                                                <PencilOutlineIcon className="icon-inline" />
                                            </Link>
                                        ))}
                                </div>
                            </dd>
                        ))}
                </dl>
                {authenticatedUser && (
                    <Link
                        to={`/users/${authenticatedUser.username}/searches`}
                        className="btn btn-secondary w-100 text-left"
                    >
                        View saved searches
                    </Link>
                )}
            </div>
        </>
    )
    const actionButtons = (
        <div className="panel-container__action-button-group">
            <div className="btn-group">
                {authenticatedUser && (
                    <Link
                        to={`/users/${authenticatedUser.username}/searches/add`}
                        className="btn btn-outline-secondary panel-container__action-button mr-2"
                    >
                        +
                    </Link>
                )}
            </div>
            <div className="btn-group">
                <button
                    type="button"
                    onClick={() => setShowAllSearches(false)}
                    className={classNames('btn btn-outline-secondary panel-container__action-button', {
                        active: !showAllSearches,
                    })}
                >
                    My searches
                </button>
                <button
                    type="button"
                    onClick={() => setShowAllSearches(true)}
                    className={classNames('btn btn-outline-secondary panel-container__action-button', {
                        active: showAllSearches,
                    })}
                >
                    All searches
                </button>
            </div>
        </div>
    )
    return (
        <PanelContainer
            className={classNames(className, 'saved-searches-panel')}
            title="Saved searches"
            state={displayState || (savedSearches ? (savedSearches.length > 0 ? 'populated' : 'empty') : 'loading')}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
