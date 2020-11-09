import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { queryUserCampaignsCodeHosts } from './backend'
import { CodeHostConnections } from './CodeHostConnections'

export interface CampaignsSettingsAreaProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryUserCampaignsCodeHosts?: typeof queryUserCampaignsCodeHosts
}

/** The page area for all campaigns settings. It's shown in the user settings sidebar. */
export const CampaignsSettingsArea: React.FunctionComponent<CampaignsSettingsAreaProps> = props => (
    <div className="web-content test-campaigns-settings-page">
        <PageTitle title="Campaigns settings" />
        <CodeHostConnections {...props} />
    </div>
)
