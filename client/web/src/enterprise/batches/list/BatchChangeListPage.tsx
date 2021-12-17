import classNames from 'classnames'
import React, { useEffect, useCallback, useState, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, ReplaySubject } from 'rxjs'
import { filter, map, tap, withLatestFrom } from 'rxjs/operators'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchChangesIcon } from '../../../batches/icons'
import { FilteredConnection, FilteredConnectionFilter } from '../../../components/FilteredConnection'
import {
    ListBatchChange,
    Scalars,
    BatchChangeState,
    BatchChangesVariables,
    BatchChangesResult,
    BatchChangesByNamespaceVariables,
} from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'

import {
    areBatchChangesLicensed as _areBatchChangesLicensed,
    queryBatchChanges as _queryBatchChanges,
    queryBatchChangesByNamespace,
} from './backend'
import styles from './BatchChangeListPage.module.scss'
import { BatchChangeNode, BatchChangeNodeProps } from './BatchChangeNode'
import { BatchChangesListIntro } from './BatchChangesListIntro'
import { GettingStarted } from './GettingStarted'
import { NewBatchChangeButton } from './NewBatchChangeButton'

export interface BatchChangeListPageProps
    extends TelemetryProps,
        Pick<RouteComponentProps, 'location'>,
        SettingsCascadeProps<Settings> {
    canCreate: boolean
    headingElement: 'h1' | 'h2'
    displayNamespace?: boolean
    /** For testing only. */
    queryBatchChanges?: typeof _queryBatchChanges
    /** For testing only. */
    areBatchChangesLicensed?: typeof _areBatchChangesLicensed
    /** For testing only. */
    openTab?: SelectedTab
}

const OPEN_FILTER = {
    label: 'Open',
    value: 'open',
    tooltip: 'Show only batch changes that are open',
    args: { state: BatchChangeState.OPEN },
} as const

const DRAFT_FILTER = {
    label: 'Draft',
    value: 'draft',
    tooltip: 'Show only batch changes that have not been applied yet',
    args: { state: BatchChangeState.DRAFT },
} as const

const CLOSED_FILTER = {
    label: 'Closed',
    value: 'closed',
    tooltip: 'Show only batch changes that are closed',
    args: { state: BatchChangeState.CLOSED },
}

const ALL_FILTER = {
    label: 'All',
    value: 'all',
    tooltip: 'Show all batch changes',
    args: {},
} as const

const getFilters = (withDrafts = false): FilteredConnectionFilter[] => [
    {
        id: 'status',
        label: 'Status',
        type: 'radio',
        values: [OPEN_FILTER, ...(withDrafts ? [DRAFT_FILTER] : []), CLOSED_FILTER, ALL_FILTER],
    },
]

type SelectedTab = 'batchChanges' | 'gettingStarted'

/**
 * A list of all batch changes on the Sourcegraph instance.
 */
export const BatchChangeListPage: React.FunctionComponent<BatchChangeListPageProps> = ({
    queryBatchChanges = _queryBatchChanges,
    areBatchChangesLicensed = _areBatchChangesLicensed,
    canCreate,
    displayNamespace = true,
    headingElement,
    location,
    openTab,
    settingsCascade,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('BatchChangesListPage'), [props.telemetryService])

    const showDrafts = isBatchChangesExecutionEnabled(settingsCascade)

    /*
     * Tracks whether this is the first fetch since this page has been rendered the first time.
     * Used to only switch to the "Getting started" tab if the user didn't select the tab manually.
     */
    const isFirstFetch = useMemo(() => {
        const subject = new ReplaySubject(1)
        subject.next(true)
        return subject
    }, [])
    const [selectedTab, setSelectedTab] = useState<SelectedTab>(openTab ?? 'batchChanges')
    const query = useCallback<(args: Partial<BatchChangesVariables>) => Observable<BatchChangesResult['batchChanges']>>(
        args =>
            queryBatchChanges(args).pipe(
                withLatestFrom(isFirstFetch),
                tap(([response, isFirst]) => {
                    if (isFirst) {
                        isFirstFetch.next(false)
                        if (!openTab && response.totalCount === 0) {
                            setSelectedTab('gettingStarted')
                        }
                    }
                }),
                // Don't emit when we are switching to the getting started tab right away to prevent a costly render.
                // Only if:
                //  - We don't fetch for the first time (the user clicked a tab) OR
                //  - There are more than 0 changesets in the namespace OR
                //  - A test forces us to display a specific tab
                filter(([response, isFirst]) => !isFirst || openTab !== undefined || response.totalCount > 0),
                map(([response]) => response.batchChanges)
            ),
        [queryBatchChanges, isFirstFetch, openTab]
    )
    const licensed: boolean | undefined = useObservable(
        useMemo(() => areBatchChangesLicensed(), [areBatchChangesLicensed])
    )

    return (
        <Page>
            <PageHeader
                path={[{ icon: BatchChangesIcon, text: 'Batch Changes' }]}
                className="test-batches-list-page mb-3"
                actions={canCreate ? <NewBatchChangeButton to={`${location.pathname}/create`} /> : null}
                headingElement={headingElement}
                description="Run custom code over hundreds of repositories and manage the resulting changesets."
            />
            <BatchChangesListIntro licensed={licensed} />
            <BatchChangeListTabHeader selectedTab={selectedTab} setSelectedTab={setSelectedTab} />
            {selectedTab === 'gettingStarted' && <GettingStarted className="mb-4" footer={<GettingStartedFooter />} />}
            {selectedTab === 'batchChanges' && (
                <Container className="mb-4">
                    <FilteredConnection<ListBatchChange, Omit<BatchChangeNodeProps, 'node'>>
                        {...props}
                        location={location}
                        nodeComponent={BatchChangeNode}
                        nodeComponentProps={{ displayNamespace }}
                        queryConnection={query}
                        hideSearch={true}
                        defaultFirst={15}
                        filters={getFilters(showDrafts)}
                        noun="batch change"
                        pluralNoun="batch changes"
                        listComponent="div"
                        listClassName={styles.batchChangeListPageGrid}
                        withCenteredSummary={true}
                        cursorPaging={true}
                        noSummaryIfAllNodesVisible={true}
                        emptyElement={<BatchChangeListEmptyElement canCreate={canCreate} location={location} />}
                    />
                </Container>
            )}
        </Page>
    )
}

export interface NamespaceBatchChangeListPageProps extends Omit<BatchChangeListPageProps, 'canCreate'> {
    authenticatedUser: AuthenticatedUser
    namespaceID: Scalars['ID']
}

/**
 * A list of all batch changes in a namespace.
 */
export const NamespaceBatchChangeListPage: React.FunctionComponent<NamespaceBatchChangeListPageProps> = ({
    authenticatedUser,
    namespaceID,
    ...props
}) => {
    // A user should only see the button to create a batch change in a namespace if it is
    // their namespace (user namespace), or they belong to it (organization namespace)
    const canCreateInThisNamespace = useMemo(
        () =>
            authenticatedUser.id === namespaceID ||
            authenticatedUser.organizations.nodes.map(org => org.id).includes(namespaceID),
        [authenticatedUser, namespaceID]
    )

    const queryConnection = useCallback(
        (args: Partial<BatchChangesByNamespaceVariables>) =>
            queryBatchChangesByNamespace({
                namespaceID,
                first: args.first ?? null,
                after: args.after ?? null,
                // The types for FilteredConnectionQueryArguments don't allow access to the filter arguments.
                state: (args as { state: BatchChangeState | undefined }).state ?? null,
                viewerCanAdminister: null,
            }),
        [namespaceID]
    )
    return (
        <BatchChangeListPage
            {...props}
            canCreate={canCreateInThisNamespace}
            displayNamespace={false}
            queryBatchChanges={queryConnection}
        />
    )
}

interface BatchChangeListEmptyElementProps extends Pick<BatchChangeListPageProps, 'location' | 'canCreate'> {}

const BatchChangeListEmptyElement: React.FunctionComponent<BatchChangeListEmptyElementProps> = ({
    canCreate,
    location,
}) => (
    <div className="w-100 py-5 text-center">
        <p>
            <strong>No batch changes have been created.</strong>
        </p>
        {canCreate ? <NewBatchChangeButton to={`${location.pathname}/create`} /> : null}
    </div>
)

const BatchChangeListTabHeader: React.FunctionComponent<{
    selectedTab: SelectedTab
    setSelectedTab: (selectedTab: SelectedTab) => void
}> = ({ selectedTab, setSelectedTab }) => {
    const onSelectBatchChanges = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('batchChanges')
        },
        [setSelectedTab]
    )
    const onSelectGettingStarted = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('gettingStarted')
        },
        [setSelectedTab]
    )
    return (
        <div className="overflow-auto mb-2">
            <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onSelectBatchChanges}
                        className={classNames('nav-link', selectedTab === 'batchChanges' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="All batch changes">
                            All batch changes
                        </span>
                    </a>
                </li>
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onSelectGettingStarted}
                        className={classNames('nav-link', selectedTab === 'gettingStarted' && 'active')}
                        role="button"
                        data-testid="test-getting-started-btn"
                    >
                        <span className="text-content" data-tab-content="Getting started">
                            Getting started
                        </span>
                    </a>
                </li>
            </ul>
        </div>
    )
}

const GettingStartedFooter: React.FunctionComponent<{}> = () => (
    <div className="row">
        <div className="col-12 col-sm-8 offset-sm-2 col-md-6 offset-md-3">
            <div className="card">
                <div className="card-body text-center">
                    <p>Create your first batch change</p>
                    <h2 className="mb-0">
                        <a href="https://docs.sourcegraph.com/batch_changes/quickstart" target="_blank" rel="noopener">
                            Batch Changes quickstart
                        </a>
                    </h2>
                </div>
            </div>
        </div>
    </div>
)
