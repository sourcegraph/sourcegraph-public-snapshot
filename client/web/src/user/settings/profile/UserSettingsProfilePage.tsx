import React, { useEffect } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { gql } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import type { EditUserProfilePage as EditUserProfilePageFragment } from '../../../graphql-operations'
import { ScimAlert } from '../ScimAlert'

import { EditUserProfileForm } from './EditUserProfileForm'

import styles from './UserSettingsProfilePage.module.scss'

export const EditUserProfilePageGQLFragment = gql`
    fragment EditUserProfilePage on User {
        id
        username
        displayName
        avatarURL
        viewerCanChangeUsername
        createdAt
        scimControlled
    }
`

interface Props extends TelemetryV2Props {
    user: EditUserProfilePageFragment
}

export const UserSettingsProfilePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('settings.profile', 'view')
        EVENT_LOGGER.logViewEvent('UserProfile')
    }, [telemetryRecorder])

    return (
        <div>
            <PageTitle title="Profile" />
            <PageHeader
                path={[{ text: 'Profile' }]}
                headingElement="h2"
                description={
                    <>
                        {user.displayName ? (
                            <>
                                {user.displayName} ({user.username})
                            </>
                        ) : (
                            user.username
                        )}{' '}
                        started using Sourcegraph <Timestamp date={user.createdAt} />.
                    </>
                }
                className={styles.heading}
            />
            {user.scimControlled && <ScimAlert />}
            {user && (
                <EditUserProfileForm
                    user={user}
                    initialValue={user}
                    telemetryRecorder={telemetryRecorder}
                    after={
                        window.context.sourcegraphDotComMode && (
                            <Text className="mt-4">
                                <Link to="mailto:support@sourcegraph.com" target="_blank" rel="noopener noreferrer">
                                    Contact support
                                </Link>{' '}
                                to delete your account.
                            </Text>
                        )
                    }
                />
            )}
        </div>
    )
}
