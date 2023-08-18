import React, { useState } from 'react'

import type { ApolloError } from '@apollo/client/errors'
import { noop } from 'lodash'
import { useNavigate } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import { Button, ErrorAlert, Form, H3, Label, Modal } from '@sourcegraph/wildcard'

import type {
    AssignOwnerResult,
    AssignOwnerVariables,
    AssignTeamResult,
    AssignTeamVariables,
    Scalars,
} from '../../graphql-operations'
import { LoaderButton } from '../LoaderButton'

import { ASSIGN_OWNER, ASSIGN_TEAM } from './graphqlQueries'
import { UserTeamSelect } from './UserTeamSelect'

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
    const [selectedTeam, setSelectedTeam] = useState<Scalars['ID']>('')
    const [error, setError] = useState<ApolloError | undefined>(undefined)
    const navigate = useNavigate()

    const [assignOwner, { loading }] = useMutation<AssignOwnerResult, AssignOwnerVariables>(ASSIGN_OWNER, {
        variables: {
            input: {
                absolutePath: path,
                assignedOwnerID: selectedUser,
                repoID,
            },
        },
        onCompleted: () => navigate(0),
        onError: error => setError(error),
    })

    const [assignTeam, { loading: teamLoading }] = useMutation<AssignTeamResult, AssignTeamVariables>(ASSIGN_TEAM, {
        variables: {
            input: {
                absolutePath: path,
                assignedOwnerID: selectedTeam,
                repoID,
            },
        },
        onCompleted: () => navigate(0),
        onError: error => setError(error),
    })

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add owner</H3>

            {error && <ErrorAlert error={error} />}

            <Form
                onSubmit={event => {
                    event.preventDefault()
                    const assignmentFunction = selectedUser !== '' ? assignOwner : assignTeam
                    assignmentFunction().catch(noop)
                }}
            >
                <Label htmlFor="add-owner--owner" className="mt-2">
                    New owner
                </Label>
                <div className="mb-3">
                    <UserTeamSelect
                        htmlID="add-owner--owner"
                        onSelectUser={user => setSelectedUser(user?.id ?? '')}
                        onSelectTeam={team => setSelectedTeam(team?.id ?? '')}
                    />
                </div>

                <div className="d-flex justify-content-end pt-1">
                    <Button
                        disabled={loading || teamLoading}
                        className="mr-2"
                        onClick={onCancel}
                        outline={true}
                        variant="secondary"
                    >
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        loading={loading || teamLoading}
                        disabled={loading || teamLoading || (selectedUser === '' && selectedTeam === '')}
                        alwaysShowLabel={true}
                        label="Add owner"
                    />
                </div>
            </Form>
        </Modal>
    )
}
