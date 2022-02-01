import React, { useEffect } from 'react'

import { gql } from '@sourcegraph/http-client'
import { percentageDone } from '@sourcegraph/shared/src/components/activation/Activation'
import { ActivationChecklist } from '@sourcegraph/shared/src/components/activation/ActivationChecklist'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { EditUserProfilePage as EditUserProfilePageFragment } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

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
    }
`

interface Props extends Pick<UserSettingsAreaRouteContext, 'activation'> {
    user: EditUserProfilePageFragment
}

export const UserSettingsProfilePage: React.FunctionComponent<Props> = ({ user, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('UserProfile'), [])

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
            {props.activation?.completed && percentageDone(props.activation.completed) < 100 && (
                <Container className="mb-3">
                    <h3>Almost there!</h3>
                    <p>Complete the steps below to finish onboarding to Sourcegraph.</p>
                    <ActivationChecklist steps={props.activation.steps} completed={props.activation.completed} />
                </Container>
            )}
            {user && (
                <EditUserProfileForm
                    user={user}
                    initialValue={user}
                    after={
                        window.context.sourcegraphDotComMode && (
                            <p className="mt-4">
                                <a href="https://about.sourcegraph.com/contact">Contact support</a> to delete your
                                account.
                            </p>
                        )
                    }
                />
            )}
        </div>
    )
}
