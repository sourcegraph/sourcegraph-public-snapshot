import React, { useState } from 'react'

import { noop } from 'lodash'
import { useNavigate } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import { Button, ErrorAlert, Form, H3, Label, Modal } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { AssignOwnerResult, AssignOwnerVariables, Scalars } from '../../graphql-operations'
import { UserSelect as SingleUserSelect } from '../../site-admin/user-select/UserSelect'

import { ASSIGN_OWNER } from './graphqlQueries'

export interface AddOwnerModalProps {
    repoID: string
    path: string
    onCancel: () => void
}

export const AddOwnerModal: React.FunctionComponent<React.PropsWithChildren<AddOwnerModalProps>> = ({
    repoID,
    path,
    onCancel,
}) => {
    const labelId = 'addOwner'
    const [selectedUser, setSelectedUser] = useState<Scalars['ID']>('')
    const navigate = useNavigate()

    const [assignOwner, { error, loading }] = useMutation<AssignOwnerResult, AssignOwnerVariables>(ASSIGN_OWNER, {
        variables: {
            input: {
                absolutePath: path,
                assignedOwnerID: selectedUser,
                repoID,
            },
        },
        onCompleted: () => navigate(0),
    })

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add owner</H3>

            {error && <ErrorAlert error={error} />}

            <Form
                onSubmit={event => {
                    event.preventDefault()
                    assignOwner().catch(noop)
                }}
            >
                <Label htmlFor="add-owner--owner" className="mt-2">
                    New owners
                </Label>
                <div className="mb-3">
                    <SingleUserSelect htmlID="add-owner--owner" onSelect={user => setSelectedUser(user?.id ?? '')} />
                </div>

                <div className="d-flex justify-content-end pt-1">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        loading={loading}
                        disabled={loading || selectedUser === ''}
                        alwaysShowLabel={true}
                        label="Add owner"
                    />
                </div>
            </Form>
        </Modal>
    )
}
