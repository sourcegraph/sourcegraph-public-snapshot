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
import { LoadingModal } from './LoadingModal'

export const SavedSearchesPanel: React.FunctionComponent<{
    patternType: SearchPatternType
    authenticatedUser: AuthenticatedUser | null
    fetchSavedSearches: () => Observable<ISavedSearch[]>
    className?: string
}> = ({ patternType, authenticatedUser, fetchSavedSearches, className }) => {
    const savedSearches = useObservable(useMemo(() => fetchSavedSearches(), [fetchSavedSearches]))
    const [showAllSearches, setShowAllSearches] = useState(true)

    const emptyDisplay = (
        <div className="panel-container__empty-container text-muted">
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
    const loadingDisplay = <LoadingModal text="Loading saved searches" />

    const contentDisplay = (
        <>
            <div className="d-flex flex-column h-100">
                <div className="d-flex justify-content-between mb-1">
                    <small>Search</small>
                    <small>Edit</small>
                </div>
                <dl className="list-group-flush flex-grow-1">
                    {savedSearches
                        ?.filter(search => (showAllSearches ? true : search.namespace.id === authenticatedUser?.id))
                        .map(search => (
                            <dd key={search.id} className="text-monospace test-saved-search-entry">
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
            <div className="btn-group btn-group-sm">
                {authenticatedUser && (
                    <Link
                        to={`/users/${authenticatedUser.username}/searches/add`}
                        className="btn btn-outline-secondary panel-container__action-button mr-2"
                    >
                        +
                    </Link>
                )}
            </div>
            <div className="btn-group btn-group-sm">
                <button
                    type="button"
                    onClick={() => setShowAllSearches(false)}
                    className={classNames(
                        'btn btn-outline-secondary panel-container__action-button test-saved-search-panel-my-searches',
                        {
                            active: !showAllSearches,
                        }
                    )}
                >
                    My searches
                </button>
                <button
                    type="button"
                    onClick={() => setShowAllSearches(true)}
                    className={classNames(
                        'btn btn-outline-secondary panel-container__action-button test-saved-search-panel-all-searches',
                        {
                            active: showAllSearches,
                        }
                    )}
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
            state={savedSearches ? (savedSearches.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
