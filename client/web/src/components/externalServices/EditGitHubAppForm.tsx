import React, { useCallback, useState } from 'react'

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
    value
    initialValue
    doUpdate
    after?: React.ReactNode
}

/**
 * A form to edit a user's profile.
 */
export const EditGitHubAppForm: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    value,
    initialValue,
    doUpdate,
    after,
}) => {
    const history = useHistory()

    const [gitHubAppFields, setGitHubAppFields] = useState(initialValue)
    const onChange = useCallback<React.ComponentProps<typeof GitHubAppFormFields>['onChange']>(newValue => {
        setGitHubAppFields(previous => ({ ...previous, ...newValue }))
    }, [])

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            setGitHubAppFields(previous => ({
                ...previous,
                gitHubApp: { ...previous.gitHubApp, privateKey: btoa(previous.gitHubApp.privateKey) },
            }))
            doUpdate(gitHubAppFields)
        },
        [doUpdate, gitHubAppFields]
    )

    return (
        <Container>
            <Form className="w-100" onSubmit={onSubmit}>
                <GitHubAppFormFields onChange={onChange} value={gitHubAppFields} disabled={false} />
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
