import React from 'react'

import { Alert, Text, Link } from '@sourcegraph/wildcard'

import { ExecutorCompatibility } from '../../../graphql-operations'

export interface ExecutorCompatibilityAlertProps {
    hostname: string
    compatibility: ExecutorCompatibility
}

export const ExecutorCompatibilityAlert: React.FunctionComponent<
    React.PropsWithChildren<ExecutorCompatibilityAlertProps>
> = ({ hostname, compatibility }) => {
    switch (compatibility) {
        case ExecutorCompatibility.OUTDATED: {
            return (
                <Alert variant="warning" className="mt-3 mb-0">
                    <Text className="m-0">{hostname} is outdated.</Text>
                    <Text className="m-0">
                        Please{' '}
                        <Link to="/help/admin/executors/deploy_executors" target="_blank" rel="noopener">
                            upgrade this executor
                        </Link>{' '}
                        to a version compatible with your Sourcegraph version.
                    </Text>
                </Alert>
            )
        }
        case ExecutorCompatibility.VERSION_AHEAD: {
            return (
                <Alert variant="warning" className="mt-3 mb-0">
                    <Text className="m-0">Your Sourcegraph instance is out of date.</Text>
                    <Text className="m-0">
                        Please{' '}
                        <Link to="/help/admin/updates" target="_blank" rel="noopener">
                            upgrade your Sourcegraph instance
                        </Link>
                        or{' '}
                        <Link to="/help/admin/executors/deploy_executors" target="_blank" rel="noopener">
                            downgrade this executor
                        </Link>
                        .
                    </Text>
                </Alert>
            )
        }
        case ExecutorCompatibility.UP_TO_DATE: {
            return null
        }
    }
}
