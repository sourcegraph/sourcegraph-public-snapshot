import React, { useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject } from 'rxjs'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsCodeHostFields } from '../../../graphql-operations'
import { CampaignsIconFlushLeft } from '../icons'
import { queryUserCampaignsCodeHosts as _queryUserCampaignsCodeHosts } from './backend'
import { CodeHostConnectionNode, CodeHostConnectionNodeProps } from './CodeHostConnectionNode'

export interface CodeHostConnectionsProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryUserCampaignsCodeHosts?: typeof _queryUserCampaignsCodeHosts
}

export const CodeHostConnections: React.FunctionComponent<CodeHostConnectionsProps> = ({
    history,
    location,
    queryUserCampaignsCodeHosts = _queryUserCampaignsCodeHosts,
}) => {
    // Subject to fire a reload of the list.
    const updateList = useMemo(() => new Subject<void>(), [])
    return (
        <>
            <PageHeader icon={CampaignsIconFlushLeft} title="Campaigns" className="justify-content-end" />
            <h2>Code host tokens</h2>
            <p>Add authentication tokens to enable campaigns changeset creation on your code hosts.</p>
            <FilteredConnection<CampaignsCodeHostFields, Omit<CodeHostConnectionNodeProps, 'node'>>
                history={history}
                location={location}
                useURLQuery={false}
                nodeComponent={CodeHostConnectionNode}
                nodeComponentProps={{ history, updateList }}
                queryConnection={queryUserCampaignsCodeHosts}
                hideSearch={true}
                defaultFirst={15}
                noun="code host"
                pluralNoun="code hosts"
                listClassName="list-group"
                updates={updateList}
                className="mb-3"
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
            />
            <p>
                Code host not present? Site admins can add a code host in{' '}
                <a href="https://docs.sourcegraph.com/admin/external_service" target="_blank" rel="noopener noreferrer">
                    the manage repositories settings
                </a>
                .
            </p>
        </>
    )
}
