import classNames from 'classnames'
import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { FilteredConnectionFilter } from '../../../components/FilteredConnection'
import { NotebooksOrderBy } from '../../../graphql-operations'
import { PageRoutes } from '../../../routes.constants'
import { fetchNotebooks as _fetchNotebooks } from '../backend'

import { SearchNotebooksList } from './SearchNotebooksList'
import styles from './SearchNotebooksListPage.module.scss'

export interface SearchNotebooksListPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    fetchNotebooks?: typeof _fetchNotebooks
}

type SelectedTab = 'my' | 'explore'

function getSelectedTabFromLocation(locationSearch: string, authenticatedUser: AuthenticatedUser | null): SelectedTab {
    if (!authenticatedUser) {
        return 'explore'
    }

    const urlParameters = new URLSearchParams(locationSearch)
    switch (urlParameters.get('tab')) {
        case 'my':
            return 'my'
        case 'explore':
            return 'explore'
    }
    return 'my'
}

function setSelectedLocationTab(location: H.Location, history: H.History, selectedTab: SelectedTab): void {
    const urlParameters = new URLSearchParams(location.search)
    urlParameters.set('tab', selectedTab)
    urlParameters.delete('visible')
    if (location.search !== urlParameters.toString()) {
        history.replace({ ...location, search: urlParameters.toString() })
    }
}

export const SearchNotebooksListPage: React.FunctionComponent<SearchNotebooksListPageProps> = ({
    authenticatedUser,
    telemetryService,
    fetchNotebooks = _fetchNotebooks,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SearchNotebooksListPage')
    }, [telemetryService])

    const history = useHistory()
    const location = useLocation()

    const [selectedTab, setSelectedTab] = useState<SelectedTab>(
        getSelectedTabFromLocation(location.search, authenticatedUser)
    )

    const setTab = useCallback(
        (tab: SelectedTab) => {
            setSelectedTab(tab)
            setSelectedLocationTab(location, history, tab)
        },
        [location, history]
    )

    const onSelectMyNotebooks = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setTab('my')
            telemetryService.log('SearchNotebooksMyNotebooksTabClick')
        },
        [setTab, telemetryService]
    )

    const onSelectExploreNotebooks = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setTab('explore')
            telemetryService.log('SearchNotebooksExploreNotebooksTabClick')
        },
        [setTab, telemetryService]
    )

    const filters: FilteredConnectionFilter[] = [
        {
            label: 'Order by',
            type: 'select',
            id: 'order',
            tooltip: 'Order notebooks',
            values: [
                {
                    value: 'updated-at-desc',
                    label: 'Last update (descending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_UPDATED_AT,
                        descending: true,
                    },
                },
                {
                    value: 'updated-at-asc',
                    label: 'Last update (ascending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_UPDATED_AT,
                        descending: false,
                    },
                },
                {
                    value: 'created-at-desc',
                    label: 'Creation date (descending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_CREATED_AT,
                        descending: true,
                    },
                },
                {
                    value: 'created-at-asc',
                    label: 'Creation date (ascending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_CREATED_AT,
                        descending: false,
                    },
                },
            ],
        },
    ]

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    path={[{ icon: MagnifyIcon, to: '/search' }, { text: 'Notebooks' }]}
                    actions={
                        authenticatedUser && (
                            <Link to={PageRoutes.NotebookCreate} className="btn btn-primary">
                                <PlusIcon className="icon-inline" />
                                Create notebook
                            </Link>
                        )
                    }
                    className="mb-3"
                />
                <div className="mb-4">
                    <div className="nav nav-tabs">
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <a
                                href=""
                                role="button"
                                onClick={onSelectMyNotebooks}
                                className={classNames('nav-link', selectedTab === 'my' && 'active')}
                            >
                                <span className="text-content" data-tab-content="My Notebooks">
                                    My Notebooks
                                </span>
                            </a>
                        </div>
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <a
                                href=""
                                role="button"
                                onClick={onSelectExploreNotebooks}
                                className={classNames('nav-link', selectedTab === 'explore' && 'active')}
                            >
                                <span className="text-content" data-tab-content="Explore Notebooks">
                                    Explore Notebooks
                                </span>
                            </a>
                        </div>
                    </div>
                </div>
                {selectedTab === 'my' && authenticatedUser && (
                    <SearchNotebooksList
                        fetchNotebooks={fetchNotebooks}
                        filters={filters}
                        authenticatedUser={authenticatedUser}
                    />
                )}
                {selectedTab === 'my' && !authenticatedUser && (
                    <UnauthenticatedMyNotebooksSection
                        telemetryService={telemetryService}
                        onSelectExploreNotebooks={onSelectExploreNotebooks}
                    />
                )}
                {selectedTab === 'explore' && <SearchNotebooksList fetchNotebooks={fetchNotebooks} filters={filters} />}
            </Page>
        </div>
    )
}

interface UnauthenticatedMyNotebooksSectionProps extends TelemetryProps {
    onSelectExploreNotebooks: (event: React.MouseEvent<HTMLElement>) => void
}

const UnauthenticatedMyNotebooksSection: React.FunctionComponent<UnauthenticatedMyNotebooksSectionProps> = ({
    telemetryService,
    onSelectExploreNotebooks,
}) => {
    const onClick = (): void => {
        telemetryService.log('SearchNotebooksSignUpToCreateNotebooksClick')
    }

    return (
        <div className="d-flex justify-content-center align-items-center flex-column p-3">
            <Link
                onClick={onClick}
                to={`/sign-up?returnTo=${encodeURIComponent('/notebooks')}`}
                className="btn btn-primary"
            >
                Sign up to create notebooks
            </Link>
            <span className="my-3 text-muted">or</span>
            <span className={classNames('d-flex align-items-center', styles.explorePublicNotebooks)}>
                <button className="btn btn-link p-1" type="button" onClick={onSelectExploreNotebooks}>
                    explore
                </button>{' '}
                public notebooks
            </span>
        </div>
    )
}
