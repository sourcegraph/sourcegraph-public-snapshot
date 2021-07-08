import React from 'react'

import { PageTitle } from '../../../components/PageTitle'

import { CodeHostConnections } from './CodeHostConnections'

/** The page area for all batch changes settings. It's shown in the site admin settings sidebar. */
export const BatchChangesSiteConfigSettingsArea: React.FunctionComponent = () => (
    <>
        <PageTitle title="Batch changes settings" />
        <CodeHostConnections
            headerLine={
                <>
                    <p>Add access tokens to enable Batch Changes changeset creation for all users.</p>
                    <div className="alert alert-info">
                        You are configuring <strong>global credentials</strong> for Batch Changes. The credentials on
                        this page can be used by all users of this Sourcegraph instance to create and sync changesets.
                    </div>
                </>
            }
            userID={null}
        />
    </>
)
