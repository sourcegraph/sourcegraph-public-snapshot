import React, { useCallback, useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { useHistory } from 'react-router'
import { UpdateUserResult, UpdateUserVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserProfileFormFields, UserProfileFormFieldsValue } from './UserProfileFormFields'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../../../branded/src/components/Form'

interface Props {
    user: Pick<GQL.IUser, 'id' | 'username' | 'viewerCanChangeUsername'>
    initialValue: UserProfileFormFieldsValue
    after?: React.ReactFragment
}

const UPDATE_USER = gql`
    mutation UpdateUser($user: ID!, $username: String!, $displayName: String, $avatarURL: String) {
        updateUser(user: $user, username: $username, displayName: $displayName, avatarURL: $avatarURL) {
            id
            username
            displayName
            avatarURL
        }
    }
`

/**
 * A form to edit a user's profile.
 */
export const EditUserProfileForm: React.FunctionComponent<Props> = ({ user, initialValue, after }) => {
    const history = useHistory()
    const [value, setValue] = useState<UserProfileFormFieldsValue>(initialValue)
    const onChange = useCallback<React.ComponentProps<typeof UserProfileFormFields>['onChange']>(
        newValue => setValue(previous => ({ ...previous, ...newValue })),
        []
    )
    const [updateUser, { data, loading, error }] = useMutation<UpdateUserResult, UpdateUserVariables>(UPDATE_USER, {
        onCompleted: ({ updateUser }) => {
            history.replace(`/users/${updateUser.username}/settings/profile`)
        },
    })

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            eventLogger.log('UpdateUserClicked')
            return updateUser({ variables: { ...value, user: user.id } })
        },
        [user.id, value, updateUser]
    )

    return (
        <Form className="w-100" onSubmit={onSubmit}>
            <UserProfileFormFields
                value={value}
                onChange={onChange}
                usernameFieldDisabled={!user.viewerCanChangeUsername}
                disabled={loading}
            />
            <button type="submit" className="btn btn-primary" disabled={loading} id="test-EditUserProfileForm__save">
                Save
            </button>
            {error && <div className="mt-3 alert alert-danger">{error.message}</div>}
            {data?.updateUser && (
                <div className="mt-3 alert alert-success test-EditUserProfileForm__success">User profile updated.</div>
            )}
            {after && (
                <>
                    <hr className="my-4" />
                    {after}
                </>
            )}
        </Form>
    )
}
