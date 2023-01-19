import React, { useCallback, useState } from 'react'

import { Button, Form, H3, Input, Modal } from '@sourcegraph/wildcard'
import { gql, useMutation } from '@sourcegraph/http-client'
import { AddGerritAccountResult, AddGerritAccountVariables } from 'src/graphql-operations'

export const ADD_GERRIT_ACCOUNT = gql`
    mutation AddGerritAccount($username: String!, $password: String!, $serviceID: String!) {
        addGerritExternalAccount(username: $username, password: $password, serviceID: $serviceID) {
            alwaysNil
        }
    }
`

export const AddGerritAccountModal: React.FunctionComponent<
    React.PropsWithChildren<{
        onDidCancel: () => void
        onDidAdd: () => void
        serviceID: string
    }>
> = ({ onDidAdd, onDidCancel, serviceID }) => {
    const [isLoading, setIsLoading] = useState(false)
    const [addGerritAccount, { error }] = useMutation<AddGerritAccountResult, AddGerritAccountVariables>(
        ADD_GERRIT_ACCOUNT
    )

    const onAccountAdd = useCallback<React.FormEventHandler<HTMLFormElement>>(async event => {
        const target = event.target as typeof event.target & {
            username: { value: string }
            password: { value: string }
        }
        event.preventDefault()
        setIsLoading(true)

        await addGerritAccount({
            variables: {
                username: target.username.value,
                password: target.password.value,
                serviceID: serviceID,
            },
        })

        if (error) {
            console.log(error)
            setIsLoading(false)
            return
        }
        setIsLoading(false)
        onDidAdd()
    }, [])

    return (
        <Modal aria-labelledby="heading--add-gerrit-account" aria-describedby="description--add-gerrit-account">
            <H3 id="heading--add-gerrit-account" className="mb-4">
                Add Gerrit account
            </H3>
            <Form onSubmit={onAccountAdd}>
                <p id="description--add-gerrit-account">
                    You are about to add a Gerrit account. Please enter your Gerrit HTTP credentials.
                </p>
                <Input type="text" name="username" placeholder="Username" className="mb-4" />
                <Input type="password" name="password" placeholder="Password" className="mb-4" />
                <div className="d-flex justify-content-end">
                    <Button
                        className="mr-2"
                        disabled={isLoading}
                        onClick={onDidCancel}
                        outline={true}
                        variant="secondary"
                    >
                        Cancel
                    </Button>
                    <Button type="submit" disabled={isLoading} variant="primary">
                        Add account
                    </Button>
                </div>
            </Form>
        </Modal>
    )
}
