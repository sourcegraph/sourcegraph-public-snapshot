import React, { useEffect } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { gql } from '@sourcegraph/http-client'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import type { EditUserProfilePage as EditUserProfilePageFragment } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
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
        eventLogger.logViewEvent('UserProfile')
        telemetryRecorder.recordEvent('userProfile', 'viewed')
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
                    after={
                        window.context.sourcegraphDotComMode && (
                            <Text className="mt-4">
                                <Link to="https://about.sourcegraph.com/contact">Contact support</Link> to delete your
                                account.
                            </Text>
                        )
                    }
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </div>
    )
}
