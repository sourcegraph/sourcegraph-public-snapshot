import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { Button, H3, Modal, ErrorAlert, Form, Label } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type { Scalars } from '../../graphql-operations'

import { useAddTeamMembers } from './backend'
import { UserSelect } from './user-select/UserSelect'

export interface AddTeamMemberModalProps {
    teamID: Scalars['ID']
    teamName: string

    onCancel: () => void
    afterAdd: () => void
}

export const AddTeamMemberModal: React.FunctionComponent<React.PropsWithChildren<AddTeamMemberModalProps>> = ({
    teamID,
    teamName,
    onCancel,
    afterAdd,
}) => {
    const labelId = 'addTeamMember'

    const [selectedMembers, setSelectedMembers] = useState<Scalars['ID'][]>([])

    const [addMembers, { loading, error }] = useAddTeamMembers()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()

            if (!event.currentTarget.checkValidity()) {
                return
            }

            try {
                await addMembers({
                    variables: { team: teamID, members: selectedMembers.map(member => ({ userID: member })) },
                })

                afterAdd()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterAdd, teamID, selectedMembers, addMembers]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add members to {teamName}</H3>

            {error && <ErrorAlert error={error} />}

            <Form onSubmit={onSubmit}>
                <Label htmlFor="add-team-member--members" className="mt-2">
                    New members
                </Label>
                <UserSelect
                    id="add-team-member--members"
                    disabled={loading}
                    setSelectedMembers={setSelectedMembers}
                    className="mb-3"
                />

                <div className="d-flex justify-content-end pt-1">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        loading={loading}
                        disabled={loading}
                        alwaysShowLabel={true}
                        label="Add members"
                    />
                </div>
            </Form>
        </Modal>
    )
}
