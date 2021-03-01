import React, { useEffect } from 'react'
import H from 'history'
import { percentageDone } from '../../../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../../../shared/src/components/activation/ActivationChecklist'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import {
    EditUserProfilePage as EditUserProfilePageFragment,
    GetUserProfileResult,
    GetUserProfileVariables,
} from '../../../graphql-operations'
import { EditUserProfileForm } from './EditUserProfileForm'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'
import { useQuery, gql } from '@apollo/client'

const GET_USER_PROFILE = gql`
    query GetUserProfile($username: String!) {
        user(username: $username) {
            id
            username
            displayName
            avatarURL
            viewerCanChangeUsername
        }
    }
`

interface Props extends Pick<UserSettingsAreaRouteContext, 'activation'> {
    user: EditUserProfilePageFragment

    history: H.History
    location: H.Location

    username: string
}

export const UserSettingsProfilePage: React.FunctionComponent<Props> = ({ username, ...props }) => {
    const { data, error } = useQuery<GetUserProfileResult, GetUserProfileVariables>(GET_USER_PROFILE, {
        variables: { username },
    })

    useEffect(() => eventLogger.logViewEvent('UserProfile'), [])

    return (
        <div className="user-settings-profile-page">
            <PageTitle title="Profile" />
            <h2>Profile</h2>

            {props.activation?.completed && percentageDone(props.activation.completed) < 100 && (
                <div className="card mb-3">
                    <div className="card-body">
                        <h3 className="mb-0">Almost there!</h3>
                        <p className="mb-0">Complete the steps below to finish onboarding to Sourcegraph.</p>
                    </div>
                    <ActivationChecklist
                        history={props.history}
                        steps={props.activation.steps}
                        completed={props.activation.completed}
                    />
                </div>
            )}
            {data?.user && !error && (
                <EditUserProfileForm
                    user={data.user}
                    initialValue={data.user}
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
