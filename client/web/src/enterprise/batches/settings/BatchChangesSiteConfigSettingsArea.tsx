import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { queryUserBatchChangesCodeHosts } from './backend'
import { CodeHostConnections } from './CodeHostConnections'

export interface BatchChangesSiteConfigSettingsAreaProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryUserBatchChangesCodeHosts?: typeof queryUserBatchChangesCodeHosts
}

/** The page area for all batch changes settings. It's shown in the site admin settings sidebar. */
export const BatchChangesSiteConfigSettingsArea: React.FunctionComponent<BatchChangesSiteConfigSettingsAreaProps> = props => (
    <div className="web-content">
        <PageTitle title="Batch changes settings" />
        <CodeHostConnections {...props} />
    </div>
)
