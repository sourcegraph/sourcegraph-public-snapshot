import React, { useEffect, useCallback, useState, useMemo } from 'react'
import { queryCampaigns as _queryCampaigns, queryCampaignsByNamespace } from './backend'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection, FilteredConnectionFilter } from '../../../components/FilteredConnection'
import { CampaignNode, CampaignNodeProps } from './CampaignNode'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    ListCampaign,
    CampaignState,
    Scalars,
    CampaignsByNamespaceVariables,
    CampaignsResult,
    CampaignsVariables,
} from '../../../graphql-operations'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from '../../../../../shared/src/components/Link'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIconFlushLeft } from '../icons'
import { CampaignsListEmpty } from './CampaignsListEmpty'
import { CampaignListIntro } from './CampaignListIntro'
import { filter, map, tap, withLatestFrom } from 'rxjs/operators'
import { Observable, ReplaySubject } from 'rxjs'
import classNames from 'classnames'

export interface CampaignListPageProps extends TelemetryProps, Pick<RouteComponentProps, 'history' | 'location'> {
    displayNamespace?: boolean
    /** For testing only. */
    queryCampaigns?: typeof _queryCampaigns
    /** For testing only. */
    openTab?: SelectedTab
}

const FILTERS: FilteredConnectionFilter[] = [
    {
        label: 'Open',
        id: 'open',
        tooltip: 'Show only campaigns that are open',
        args: { state: CampaignState.OPEN },
    },
    {
        label: 'Closed',
        id: 'closed',
        tooltip: 'Show only campaigns that are closed',
        args: { state: CampaignState.CLOSED },
    },
    {
        label: 'All',
        id: 'all',
        tooltip: 'Show all campaigns',
        args: {},
    },
]

type SelectedTab = 'campaigns' | 'gettingStarted'

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const CampaignListPage: React.FunctionComponent<CampaignListPageProps> = ({
    queryCampaigns = _queryCampaigns,
    displayNamespace = true,
    location,
    openTab,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('CampaignsListPage'), [props.telemetryService])

    /*
     * Tracks whether this is the first fetch since this page has been rendered the first time.
     * Used to only switch to the "Getting started" tab if the user didn't select the tab manually.
     */
    const isFirstFetch = useMemo(() => {
        const subject = new ReplaySubject(1)
        subject.next(true)
        return subject
    }, [])
    const [selectedTab, setSelectedTab] = useState<SelectedTab>(openTab ?? 'campaigns')
    const query = useCallback<(args: Partial<CampaignsVariables>) => Observable<CampaignsResult['campaigns']>>(
        args =>
            queryCampaigns(args).pipe(
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
                map(([response]) => response.campaigns)
            ),
        [queryCampaigns, isFirstFetch, openTab]
    )

    return (
        <>
            <PageHeader
                icon={CampaignsIconFlushLeft}
                title="Campaigns"
                className="justify-content-end test-campaign-list-page"
                actions={<NewCampaignButton location={location} />}
            />
            <p className="text-muted">
                Run custom code over hundreds of repositories and manage the resulting changesets
            </p>
            <CampaignListIntro />
            <CampaignListTabHeader selectedTab={selectedTab} setSelectedTab={setSelectedTab} />
            {selectedTab === 'gettingStarted' && <CampaignsListEmpty />}
            {selectedTab === 'campaigns' && (
                <FilteredConnection<ListCampaign, Omit<CampaignNodeProps, 'node'>>
                    {...props}
                    location={location}
                    nodeComponent={CampaignNode}
                    nodeComponentProps={{ history: props.history, displayNamespace }}
                    queryConnection={query}
                    hideSearch={true}
                    defaultFirst={15}
                    filters={FILTERS}
                    noun="campaign"
                    pluralNoun="campaigns"
                    listComponent="div"
                    listClassName="campaign-list-page__grid mb-3"
                    className="mb-3"
                    cursorPaging={true}
                    noSummaryIfAllNodesVisible={true}
                    emptyElement={<CampaignListEmptyElement location={location} />}
                />
            )}
        </>
    )
}

export interface NamespaceCampaignListPageProps extends CampaignListPageProps {
    namespaceID: Scalars['ID']
}

/**
 * A list of all campaigns in a namespace.
 */
export const NamespaceCampaignListPage: React.FunctionComponent<NamespaceCampaignListPageProps> = ({
    namespaceID,
    ...props
}) => {
    const queryConnection = useCallback(
        (args: Partial<CampaignsByNamespaceVariables>) =>
            queryCampaignsByNamespace({
                namespaceID,
                first: args.first ?? null,
                after: args.after ?? null,
                // The types for FilteredConnectionQueryArguments don't allow access to the filter arguments.
                state: (args as { state: CampaignState | undefined }).state ?? null,
                viewerCanAdminister: null,
            }),
        [namespaceID]
    )
    return <CampaignListPage {...props} displayNamespace={false} queryCampaigns={queryConnection} />
}

interface CampaignListEmptyElementProps extends Pick<RouteComponentProps, 'location'> {}

const CampaignListEmptyElement: React.FunctionComponent<CampaignListEmptyElementProps> = ({ location }) => (
    <div className="w-100 py-5 text-center">
        <p>
            <strong>No campaigns have been created</strong>
        </p>
        <NewCampaignButton location={location} />
    </div>
)

interface NewCampaignButtonProps extends Pick<RouteComponentProps, 'location'> {}

const NewCampaignButton: React.FunctionComponent<NewCampaignButtonProps> = ({ location }) => (
    <Link to={`${location.pathname}/create`} className="btn btn-primary">
        <PlusIcon className="icon-inline" /> New campaign
    </Link>
)

const CampaignListTabHeader: React.FunctionComponent<{
    selectedTab: SelectedTab
    setSelectedTab: (selectedTab: SelectedTab) => void
}> = ({ selectedTab, setSelectedTab }) => {
    const onSelectCampaigns = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('campaigns')
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
        <div className="overflow-auto mb-4">
            <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
                <li className="nav-item">
                    <a
                        href=""
                        onClick={onSelectCampaigns}
                        className={classNames('nav-link', selectedTab === 'campaigns' && 'active')}
                    >
                        All campaigns
                    </a>
                </li>
                <li className="nav-item">
                    <a
                        href=""
                        onClick={onSelectGettingStarted}
                        className={classNames('nav-link', selectedTab === 'gettingStarted' && 'active')}
                    >
                        Getting started
                    </a>
                </li>
            </ul>
        </div>
    )
}
