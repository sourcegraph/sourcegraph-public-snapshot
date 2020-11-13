import React, { useEffect, useCallback, useState } from 'react'
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
import { filter, map, tap } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { SourcegraphIcon } from '../../../auth/icons'

export interface CampaignListPageProps extends TelemetryProps, Pick<RouteComponentProps, 'history' | 'location'> {
    displayNamespace?: boolean
    queryCampaigns?: typeof _queryCampaigns
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

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const CampaignListPage: React.FunctionComponent<CampaignListPageProps> = ({
    queryCampaigns = _queryCampaigns,
    displayNamespace = true,
    location,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('CampaignsListPage'), [props.telemetryService])
    const [totalCampaignsCount, setTotalCampaignsCount] = useState<number>()
    const query = useCallback<(args: Partial<CampaignsVariables>) => Observable<CampaignsResult['campaigns']>>(
        args =>
            queryCampaigns(args).pipe(
                tap(response => {
                    setTotalCampaignsCount(response.totalCount)
                }),
                filter(response => response.totalCount > 0),
                map(response => response.campaigns)
            ),
        [queryCampaigns]
    )
    return (
        <>
            <PageHeader
                icon={CampaignsIconFlushLeft}
                title="Campaigns"
                className="justify-content-end test-campaign-list-page"
                actions={
                    <Link to={`${location.pathname}/create`} className="btn btn-primary">
                        <PlusIcon className="icon-inline" /> New campaign
                    </Link>
                }
            />
            <p className="text-muted">
                Run custom code over hundreds of repositories and manage the resulting changesets
            </p>
            <div className="row">
                <div className="col-12 col-md-6 mb-2">
                    <div className="campaign-list-page__intro-card card p-2 h-100">
                        <div className="card-body d-flex align-items-start">
                            <SourcegraphIcon className="mr-3 col-2 mt-2" />
                            <div>
                                <h4>Campaigns trial</h4>
                                <p className="text-muted mb-0">
                                    Campaigns will be a paid feature in a future release. In the meantime, we invite you
                                    to trial the ability to make large scale changes across many repositories and code
                                    hosts. If you’d like to discuss use cases and features,{' '}
                                    <a href="https://about.sourcegraph.com/contact/sales/">please get in touch</a>!
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="col-12 col-md-6 mb-2">
                    <div className="campaign-list-page__intro-card card h-100 p-2">
                        <div className="card-body">
                            <h4>New campaigns features in version 3.22</h4>
                            <ul className="text-muted mb-0 pl-3">
                                <li>All users can now create campaigns</li>
                                <li>
                                    Publishing a changeset requires a code host token from the user applying the
                                    campaign
                                </li>
                                <li>
                                    Template variables such as <code>search_result_paths</code> and{' '}
                                    <code>modified_files</code> are now available in campaign specifications
                                </li>
                                <li>Rate limit improvements</li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
            {totalCampaignsCount === 0 && <CampaignsListEmpty />}
            {totalCampaignsCount !== 0 && (
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
