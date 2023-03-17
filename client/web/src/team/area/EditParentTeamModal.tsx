import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { Button, ErrorAlert, Form, H3, Input, Label, Modal } from '@sourcegraph/wildcard'

import { TEAM_DISPLAY_NAME_MAX_LENGTH } from '..'
import { LoaderButton } from '../../components/LoaderButton'
import { Scalars } from '../../graphql-operations'

import { useAssignParentTeam } from './backend'

interface EditParentTeamModalProps {
    teamID: Scalars['ID']
    teamName: string
    parentTeamName: string | null

    onCancel: () => void
    afterEdit: () => void
}

export const EditParentTeamModal: React.FunctionComponent<React.PropsWithChildren<EditParentTeamModalProps>> = ({
    teamID,
    teamName,
    parentTeamName: currentParentTeamName,
    onCancel,
    afterEdit,
}) => {
    const labelId = 'editParentTeam'

    const [parentTeamName, setParentTeamName] = useState<string | null>(currentParentTeamName)
    const onParentTeamNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setParentTeamName(event.currentTarget.value)
    }

    const [editTeam, { loading, error }] = useAssignParentTeam()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            if (!event.currentTarget.checkValidity()) {
                return
            }
            try {
                await editTeam({ variables: { id: teamID, parentTeamName } })
                afterEdit()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterEdit, teamID, parentTeamName, editTeam]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Assign parent team of {teamName}</H3>
            {error && <ErrorAlert error={error} />}
            <Form onSubmit={onSubmit}>
                <Label htmlFor="edit-team--parent" className="mt-2">
                    Display name
                </Label>
                <Input
                    id="edit-team--parent"
                    placeholder="Engineering Team"
                    maxLength={TEAM_DISPLAY_NAME_MAX_LENGTH}
                    autoCorrect="off"
                    value={parentTeamName ?? ''}
                    onChange={onParentTeamNameChange}
                    disabled={loading}
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
                        label="Save"
                    />
                </div>
            </Form>
        </Modal>
    )
}
