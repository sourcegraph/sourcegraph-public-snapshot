import H from 'history'
import React, { useCallback, useEffect } from 'react'

import { percentageDone } from '@sourcegraph/shared/src/components/activation/Activation'
import { ActivationChecklist } from '@sourcegraph/shared/src/components/activation/ActivationChecklist'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { refreshAuthenticatedUser } from '../../../auth'
import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { EditUserProfilePage as EditUserProfilePageFragment } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { EditUserProfileForm } from './EditUserProfileForm'

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

interface Props extends Pick<UserSettingsAreaRouteContext, 'onUserUpdate' | 'activation' | 'authenticatedUser'> {
    user: EditUserProfilePageFragment

    history: H.History
    location: H.Location
}

export const UserSettingsProfilePage: React.FunctionComponent<Props> = ({
    user,
    authenticatedUser,
    onUserUpdate: parentOnUpdate,
    ...props
}) => {
    useEffect(() => eventLogger.logViewEvent('UserProfile'), [])

    const onUpdate = useCallback<React.ComponentProps<typeof EditUserProfileForm>['onUpdate']>(
        newValue => {
            // Handle when username changes.
            if (newValue.username !== user.username) {
                props.history.push(`/users/${newValue.username}/settings/profile`)
            }

            parentOnUpdate(newValue)

            // In case the edited user is the current user, immediately reflect the changes in the
            // UI.
            refreshAuthenticatedUser()
                .toPromise()
                .finally(() => {})
        },
        [parentOnUpdate, props.history, user.username]
    )

    return (
        <div className="user-settings-profile-page">
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
                className="user-settings-profile-page__heading"
            />
            {props.activation?.completed && percentageDone(props.activation.completed) < 100 && (
                <Container className="mb-3">
                    <h3>Almost there!</h3>
                    <p>Complete the steps below to finish onboarding to Sourcegraph.</p>
                    <ActivationChecklist
                        history={props.history}
                        steps={props.activation.steps}
                        completed={props.activation.completed}
                    />
                </Container>
            )}
            {user && !isErrorLike(user) && (
                <EditUserProfileForm
                    user={user}
                    authenticatedUser={authenticatedUser}
                    initialValue={user}
                    onUpdate={onUpdate}
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
