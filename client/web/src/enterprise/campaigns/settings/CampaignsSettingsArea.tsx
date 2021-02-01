import React from 'react'
import { RouteComponentProps } from 'react-router'
import { Page } from '../../../components/Page'
import { PageHeader } from '../../../components/PageHeader'
import { PageTitle } from '../../../components/PageTitle'
import { UserAreaUserFields } from '../../../graphql-operations'
import { CampaignsIcon } from '../icons'
import { queryUserCampaignsCodeHosts } from './backend'
import { CodeHostConnections } from './CodeHostConnections'

export interface CampaignsSettingsAreaProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    user: Pick<UserAreaUserFields, 'id'>
    queryUserCampaignsCodeHosts?: typeof queryUserCampaignsCodeHosts
}

/** The page area for all campaigns settings. It's shown in the user settings sidebar. */
export const CampaignsSettingsArea: React.FunctionComponent<CampaignsSettingsAreaProps> = props => (
    <Page className="test-campaigns-settings-page">
        <PageTitle title="Campaigns settings" />
        <PageHeader path={[{ icon: CampaignsIcon, text: 'Campaigns' }]} className="mb-3" />
        <CodeHostConnections userID={props.user.id} {...props} />
    </Page>
)
