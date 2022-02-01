import classNames from 'classnames'
import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader, Link, Button } from '@sourcegraph/wildcard'

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

type SelectedTab = 'my' | 'explore' | 'starred'

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
        case 'starred':
            return 'starred'
    }
    return 'my'
}

function setSelectedLocationTab(location: H.Location, history: H.History, selectedTab: SelectedTab): void {
    const urlParameters = new URLSearchParams(location.search)
    urlParameters.set('tab', selectedTab)
    // Reset FilteredConnection URL params when switching between tabs
    urlParameters.delete('visible')
    urlParameters.delete('query')
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

    const onSelectTab = useCallback(
        (tab: SelectedTab, logName: string) => {
            setTab(tab)
            telemetryService.log(logName)
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
                    value: 'stars-desc',
                    label: 'Stars (descending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_STAR_COUNT,
                        descending: true,
                    },
                },
                {
                    value: 'stars-asc',
                    label: 'Stars (ascending)',
                    args: {
                        orderBy: NotebooksOrderBy.NOTEBOOK_STAR_COUNT,
                        descending: false,
                    },
                },
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
                            <Button to={PageRoutes.NotebookCreate} variant="primary" as={Link}>
                                <PlusIcon className="icon-inline" />
                                Create notebook
                            </Button>
                        )
                    }
                    className="mb-3"
                />
                <div className="mb-4">
                    <div className="nav nav-tabs">
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <Link
                                to=""
                                role="button"
                                onClick={event => {
                                    event.preventDefault()
                                    onSelectTab('my', 'SearchNotebooksMyNotebooksTabClick')
                                }}
                                className={classNames('nav-link', selectedTab === 'my' && 'active')}
                            >
                                <span className="text-content" data-tab-content="My Notebooks">
                                    My Notebooks
                                </span>
                            </Link>
                        </div>
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <Link
                                to=""
                                role="button"
                                onClick={event => {
                                    event.preventDefault()
                                    onSelectTab('starred', 'SearchNotebooksStarredNotebooksTabClick')
                                }}
                                className={classNames('nav-link', selectedTab === 'starred' && 'active')}
                            >
                                <span className="text-content" data-tab-content="Starred Notebooks">
                                    Starred Notebooks
                                </span>
                            </Link>
                        </div>
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <Link
                                to=""
                                role="button"
                                onClick={event => {
                                    event.preventDefault()
                                    onSelectTab('explore', 'SearchNotebooksExploreNotebooksTabClick')
                                }}
                                className={classNames('nav-link', selectedTab === 'explore' && 'active')}
                            >
                                <span className="text-content" data-tab-content="Explore Notebooks">
                                    Explore Notebooks
                                </span>
                            </Link>
                        </div>
                    </div>
                </div>
                {selectedTab === 'my' && authenticatedUser && (
                    <SearchNotebooksList
                        logEventName="MyNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        filters={filters}
                        creatorUserID={authenticatedUser.id}
                        telemetryService={telemetryService}
                    />
                )}
                {selectedTab === 'starred' && authenticatedUser && (
                    <SearchNotebooksList
                        logEventName="StarredNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        starredByUserID={authenticatedUser.id}
                        filters={filters}
                        telemetryService={telemetryService}
                    />
                )}
                {(selectedTab === 'my' || selectedTab === 'starred') && !authenticatedUser && (
                    <UnauthenticatedNotebooksSection
                        cta={selectedTab === 'my' ? 'Get started creating notebooks' : 'Get started starring notebooks'}
                        telemetryService={telemetryService}
                        onSelectExploreNotebooks={() =>
                            onSelectTab('explore', 'SearchNotebooksExploreNotebooksTabClick')
                        }
                    />
                )}
                {selectedTab === 'explore' && (
                    <SearchNotebooksList
                        logEventName="ExploreNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        filters={filters}
                        telemetryService={telemetryService}
                    />
                )}
            </Page>
        </div>
    )
}

interface UnauthenticatedMyNotebooksSectionProps extends TelemetryProps {
    cta: string
    onSelectExploreNotebooks: () => void
}

const UnauthenticatedNotebooksSection: React.FunctionComponent<UnauthenticatedMyNotebooksSectionProps> = ({
    telemetryService,
    cta,
    onSelectExploreNotebooks,
}) => {
    const onClick = (): void => {
        telemetryService.log('SearchNotebooksSignUpToCreateNotebooksClick')
    }

    return (
        <div className="d-flex justify-content-center align-items-center flex-column p-3">
            <Button
                as={Link}
                onClick={onClick}
                to={buildGetStartedURL('search-notebooks', '/notebooks')}
                variant="primary"
            >
                {cta}
            </Button>
            <span className="my-3 text-muted">or</span>
            <span className={classNames('d-flex align-items-center', styles.explorePublicNotebooks)}>
                <Button className="p-1" variant="link" onClick={onSelectExploreNotebooks}>
                    explore
                </Button>{' '}
                public notebooks
            </span>
        </div>
    )
}
