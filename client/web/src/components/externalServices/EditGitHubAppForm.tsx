import React, { useCallback } from 'react'

import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql } from '@sourcegraph/http-client'
import { Container, Button } from '@sourcegraph/wildcard'

import { GitHubAppFormFields } from './GitHubAppFormFields'

export const UPDATE_USER = gql`
    mutation UpdateUser($user: ID!, $username: String!, $displayName: String, $avatarURL: String) {
        updateUser(user: $user, username: $username, displayName: $displayName, avatarURL: $avatarURL) {
            id
            username
            displayName
            avatarURL
        }
    }
`

interface Props {
    after?: React.ReactNode
}

/**
 * A form to edit a user's profile.
 */
export const EditGitHubAppForm: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ after }) => {
    const history = useHistory()

    const onSubmit = useCallback<React.FormEventHandler>(event => {}, [])

    return (
        <Container>
            <Form className="w-100" onSubmit={onSubmit}>
                <GitHubAppFormFields disabled={false} />
                <Button type="submit" disabled={false} id="test-EditUserProfileForm__save" variant="primary">
                    Save
                </Button>
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
