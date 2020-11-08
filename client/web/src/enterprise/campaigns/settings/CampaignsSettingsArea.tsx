import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { queryUserCampaignsCodeHosts } from './backend'
import { CodeHostConnections } from './CodeHostConnections'

export interface CampaignsSettingsAreaProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryUserCampaignsCodeHosts?: typeof queryUserCampaignsCodeHosts
}

export const CampaignsSettingsArea: React.FunctionComponent<CampaignsSettingsAreaProps> = props => (
    <>
        <PageTitle title="Campaigns settings" />
        <CodeHostConnections {...props} />
    </>
)
