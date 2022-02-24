import classNames from 'classnames'
import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, useHistory, useLocation } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, startWith, switchMap } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader, Link, Button, useEventObservable, Alert, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { FilteredConnectionFilter } from '../../components/FilteredConnection'
import { CreateNotebookVariables, NotebooksOrderBy } from '../../graphql-operations'
import { PageRoutes } from '../../routes.constants'
import { fetchNotebooks as _fetchNotebooks, createNotebook as _createNotebook } from '../backend'

import { ImportMarkdownNotebookButton } from './ImportMarkdownNotebookButton'
import { NotebooksList } from './NotebooksList'
import styles from './NotebooksListPage.module.scss'

export interface NotebooksListPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    fetchNotebooks?: typeof _fetchNotebooks
    createNotebook?: typeof _createNotebook
}

type NotebooksTab =
    | { type: 'my' }
    | { type: 'explore' }
    | { type: 'starred' }
    | { type: 'org'; name: string; id: string }

type Tabs = { tab: NotebooksTab; title: string; isActive: boolean; logName: string }[]

function getSelectedTabFromLocation(locationSearch: string, authenticatedUser: AuthenticatedUser | null): NotebooksTab {
    if (!authenticatedUser) {
        return { type: 'explore' }
    }

    const urlParameters = new URLSearchParams(locationSearch)
    switch (urlParameters.get('tab')) {
        case 'my':
            return { type: 'my' }
        case 'explore':
            return { type: 'explore' }
        case 'starred':
            return { type: 'starred' }
    }

    const orgName = urlParameters.get('org')
    const org = orgName && authenticatedUser.organizations.nodes.find(org => org.name === orgName)
    if (org) {
        return { type: 'org', name: org.name, id: org.id }
    }

    return { type: 'my' }
}

function setSelectedLocationTab(location: H.Location, history: H.History, selectedTab: NotebooksTab): void {
    const urlParameters = new URLSearchParams(location.search)
    // Reset FilteredConnection URL params when switching between tabs
    for (const parameter of ['visible', 'query', 'order', 'org', 'tab']) {
        urlParameters.delete(parameter)
    }

    if (selectedTab.type === 'org') {
        urlParameters.set('org', selectedTab.name)
    } else {
        urlParameters.set('tab', selectedTab.type)
    }
    if (location.search !== urlParameters.toString()) {
        history.replace({ ...location, search: urlParameters.toString() })
    }
}

const LOADING = 'loading' as const

export const NotebooksListPage: React.FunctionComponent<NotebooksListPageProps> = ({
    authenticatedUser,
    telemetryService,
    fetchNotebooks = _fetchNotebooks,
    createNotebook = _createNotebook,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SearchNotebooksListPage')
    }, [telemetryService])

    const [importState, setImportState] = useState<typeof LOADING | ErrorLike | undefined>()
    const history = useHistory()
    const location = useLocation()

    const [selectedTab, setSelectedTab] = useState<NotebooksTab>(
        getSelectedTabFromLocation(location.search, authenticatedUser)
    )

    const onSelectTab = useCallback(
        (tab: NotebooksTab, logName: string) => {
            setSelectedTab(tab)
            setSelectedLocationTab(location, history, tab)
            telemetryService.log(logName)
        },
        [history, location, setSelectedTab, telemetryService]
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

    const orgTabs: Tabs | undefined = useMemo(
        () =>
            authenticatedUser?.organizations.nodes.map(org => ({
                tab: { type: 'org', name: org.name, id: org.id },
                title: `${org.name} Notebooks`,
                isActive: selectedTab.type === 'org' && selectedTab.id === org.id,
                logName: 'OrgNotebooks',
            })),
        [authenticatedUser, selectedTab]
    )

    const tabs: Tabs = useMemo(
        () => [
            {
                tab: { type: 'my' },
                title: 'My Notebooks',
                isActive: selectedTab.type === 'my',
                logName: 'MyNotebooks',
            },
            {
                tab: { type: 'starred' },
                title: 'Starred Notebooks',
                isActive: selectedTab.type === 'starred',
                logName: 'StarredNotebooks',
            },
            ...(orgTabs ?? []),
            {
                tab: { type: 'explore' },
                title: 'Explore Notebooks',
                isActive: selectedTab.type === 'explore',
                logName: 'ExploreNotebooks',
            },
        ],
        [selectedTab, orgTabs]
    )

    const [importNotebook, importedNotebookOrError] = useEventObservable(
        useCallback(
            (notebook: Observable<CreateNotebookVariables['notebook']>) =>
                notebook.pipe(
                    switchMap(notebook =>
                        createNotebook({ notebook }).pipe(
                            startWith(LOADING),
                            catchError(error => {
                                setImportState(asError(error))
                                return []
                            })
                        )
                    )
                ),
            [createNotebook, setImportState]
        )
    )

    if (importedNotebookOrError && importedNotebookOrError !== LOADING) {
        telemetryService.log('SearchNotebookImportedFromMarkdown')
        return <Redirect to={PageRoutes.Notebook.replace(':id', importedNotebookOrError.id)} />
    }

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    path={[{ icon: MagnifyIcon, to: '/search' }, { text: 'Notebooks' }]}
                    actions={
                        authenticatedUser && (
                            <>
                                <Button to={PageRoutes.NotebookCreate} variant="primary" as={Link} className="mr-2">
                                    <Icon className="mr-1" as={PlusIcon} />
                                    Create notebook
                                </Button>
                                <ImportMarkdownNotebookButton
                                    telemetryService={telemetryService}
                                    authenticatedUser={authenticatedUser}
                                    importNotebook={importNotebook}
                                    importState={importState}
                                    setImportState={setImportState}
                                />
                            </>
                        )
                    }
                    className="mb-3"
                />
                {isErrorLike(importState) && (
                    <Alert variant="danger">
                        Error while importing the notebook: <strong>{importState.message}</strong>
                    </Alert>
                )}
                <div className="mb-4">
                    <div className="nav nav-tabs">
                        {tabs.map(({ tab, title, isActive, logName }) => (
                            <div className="nav-item" key={`${tab.type}-${tab.type === 'org' && tab.id}`}>
                                <Link
                                    to=""
                                    role="button"
                                    onClick={event => {
                                        event.preventDefault()
                                        onSelectTab(tab, `SearchNotebooks${logName}TabClick`)
                                    }}
                                    className={classNames('nav-link', isActive && 'active')}
                                >
                                    <span className="text-content" data-tab-content={title}>
                                        {title}
                                    </span>
                                </Link>
                            </div>
                        ))}
                    </div>
                </div>
                {selectedTab.type === 'my' && authenticatedUser && (
                    <NotebooksList
                        logEventName="MyNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        filters={filters}
                        creatorUserID={authenticatedUser.id}
                        telemetryService={telemetryService}
                    />
                )}
                {selectedTab.type === 'starred' && authenticatedUser && (
                    <NotebooksList
                        logEventName="StarredNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        starredByUserID={authenticatedUser.id}
                        filters={filters}
                        telemetryService={telemetryService}
                    />
                )}
                {selectedTab.type === 'org' && (
                    <NotebooksList
                        logEventName="OrgNotebooks"
                        fetchNotebooks={fetchNotebooks}
                        namespace={selectedTab.id}
                        filters={filters}
                        telemetryService={telemetryService}
                    />
                )}
                {(selectedTab.type === 'my' || selectedTab.type === 'starred') && !authenticatedUser && (
                    <UnauthenticatedNotebooksSection
                        cta={
                            selectedTab.type === 'my'
                                ? 'Get started creating notebooks'
                                : 'Get started starring notebooks'
                        }
                        telemetryService={telemetryService}
                        onSelectExploreNotebooks={() =>
                            onSelectTab({ type: 'explore' }, 'SearchNotebooksExploreNotebooksTabClick')
                        }
                    />
                )}
                {selectedTab.type === 'explore' && (
                    <NotebooksList
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
