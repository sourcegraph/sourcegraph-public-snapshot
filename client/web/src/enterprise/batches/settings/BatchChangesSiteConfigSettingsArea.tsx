import React from 'react'

import { PageHeader, Alert, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'

import { GlobalCodeHostConnections } from './CodeHostConnections'
import { RolloutWindowsConfiguration } from './RolloutWindowsConfiguration'

export interface BatchChangesSiteConfigSettingsAreaProps {}

/** The page area for all batch changes settings. It's shown in the site admin settings sidebar. */
export const BatchChangesSiteConfigSettingsArea: React.FunctionComponent<
    React.PropsWithChildren<BatchChangesSiteConfigSettingsAreaProps>
> = () => (
    <>
        <PageTitle title="Batch changes settings" />
        <PageHeader headingElement="h2" path={[{ text: 'Batch Changes settings' }]} className="mb-3" />
        <RolloutWindowsConfiguration />
        <GlobalCodeHostConnections
            headerLine={
                <>
                    <Text>Add access tokens to enable Batch Changes changeset creation for all users.</Text>
                    <Alert variant="info">
                        You are configuring <strong>global credentials</strong> for Batch Changes. The credentials on
                        this page can be used by all users of this Sourcegraph instance to create and sync changesets.
                    </Alert>
                </>
            }
        />
    </>
)
