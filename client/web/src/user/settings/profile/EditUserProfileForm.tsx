import React, { useCallback, useState } from 'react'
import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { Container } from '@sourcegraph/wildcard'

import { useUpdateUser } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import { UserProfileFormFields, UserProfileFormFieldsValue } from './UserProfileFormFields'

interface Props {
    user: Pick<GQL.IUser, 'id' | 'viewerCanChangeUsername'>
    initialValue: UserProfileFormFieldsValue
    after?: React.ReactFragment
}

/**
 * A form to edit a user's profile.
 */
export const EditUserProfileForm: React.FunctionComponent<Props> = ({ user, initialValue, after }) => {
    const history = useHistory()
    const [updateUser, { data, loading, error }] = useUpdateUser({
        onCompleted: ({ updateUser }) => {
            history.replace(`/users/${updateUser.username}/settings/profile`)
        },
    })

    const [userFields, setUserFields] = useState<UserProfileFormFieldsValue>(initialValue)
    const onChange = useCallback<React.ComponentProps<typeof UserProfileFormFields>['onChange']>(
        newValue => setUserFields(previous => ({ ...previous, ...newValue })),
        []
    )

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            eventLogger.log('UpdateUserClicked')
            return updateUser({
                variables: {
                    user: user.id,
                    username: userFields.username,
                    displayName: userFields.displayName,
                    avatarURL: userFields.avatarURL,
                },
            })
        },
        [updateUser, user.id, userFields]
    )

    return (
        <Container>
            <Form className="w-100" onSubmit={onSubmit}>
                <UserProfileFormFields
                    value={userFields}
                    onChange={onChange}
                    usernameFieldDisabled={!user.viewerCanChangeUsername}
                    disabled={loading}
                />
                <button
                    type="submit"
                    className="btn btn-primary"
                    disabled={loading}
                    id="test-EditUserProfileForm__save"
                >
                    Save
                </button>
                {error && <div className="mt-3 alert alert-danger">{error.message}</div>}
                {data?.updateUser && (
                    <div className="mt-3 mb-0 alert alert-success test-EditUserProfileForm__success">
                        User profile updated.
                    </div>
                )}
                {after && (
                    <>
                        <hr className="my-4" />
                        {after}
                    </>
                )}
            </Form>
        </Container>
    )
}
