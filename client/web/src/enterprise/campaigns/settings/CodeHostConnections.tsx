import React from 'react'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsCodeHostFields } from '../../../graphql-operations'
import { CampaignsIconFlushLeft } from '../icons'
import { queryUserCampaignsCodeHosts } from './backend'
import { CodeHostConnectionNode, CodeHostConnectionNodeProps } from './CodeHostConnectionNode'

export interface CodeHostConnectionsProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    // Nothing for now.
}

export const CodeHostConnections: React.FunctionComponent<CodeHostConnectionsProps> = props => (
    <>
        <PageHeader icon={CampaignsIconFlushLeft} title="Campaigns" className="justify-content-end" />
        <h2>Code host connections</h2>
        <p>Configure access to code hosts connected to this Sourcegraph instance, so you can publish changesets.</p>
        <FilteredConnection<CampaignsCodeHostFields, Omit<CodeHostConnectionNodeProps, 'node'>>
            {...props}
            nodeComponent={CodeHostConnectionNode}
            nodeComponentProps={{ history: props.history }}
            queryConnection={queryUserCampaignsCodeHosts}
            hideSearch={true}
            defaultFirst={15}
            noun="code host"
            pluralNoun="code hosts"
            listClassName="list-group"
            className="mb-3"
            cursorPaging={true}
            noSummaryIfAllNodesVisible={true}
        />
    </>
)
