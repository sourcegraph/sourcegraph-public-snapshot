import InfoCircleIcon from 'mdi-react/InfoCircleIcon'
import React from 'react'
import { RouteComponentProps } from 'react-router'

import { PageTitle } from '../../../components/PageTitle'

import { queryGlobalBatchChangesCodeHosts } from './backend'
import { CodeHostConnections } from './CodeHostConnections'

export interface BatchChangesSiteConfigSettingsAreaProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryGlobalBatchChangesCodeHosts?: typeof queryGlobalBatchChangesCodeHosts
}

/** The page area for all batch changes settings. It's shown in the site admin settings sidebar. */
export const BatchChangesSiteConfigSettingsArea: React.FunctionComponent<BatchChangesSiteConfigSettingsAreaProps> = props => (
    <div className="web-content">
        <PageTitle title="Batch changes settings" />
        <CodeHostConnections
            headerLine={
                <>
                    <p>Add access tokens to enable Batch Changes changeset creation for all users.</p>
                    <div className="alert alert-info">
                        <InfoCircleIcon className="redesign-d-none icon-inline mr-2" />
                        You are configuring <strong>global credentials</strong> for Batch Changes. The credentials on
                        this page can be used by all users of this Sourcegraph instance to create and sync changesets.
                    </div>
                </>
            }
            userID={null}
            {...props}
        />
    </div>
)
