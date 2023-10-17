import React, { useCallback } from 'react'

import { logger } from '@sourcegraph/common'
import { Button, H3, Modal, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type { ListTeamFields } from '../../graphql-operations'

import { useDeleteTeam } from './backend'

export interface DeleteTeamModalProps {
    team: ListTeamFields

    onCancel: () => void
    afterDelete: () => void
}

export const DeleteTeamModal: React.FunctionComponent<React.PropsWithChildren<DeleteTeamModalProps>> = ({
    team,
    onCancel,
    afterDelete,
}) => {
    const labelId = 'deleteTeam'

    const [deleteTeam, { loading, error }] = useDeleteTeam()

    const onDelete = useCallback<React.MouseEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await deleteTeam({ variables: { id: team.id } })

                afterDelete()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterDelete, team.id, deleteTeam]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Delete team {team.name}?</H3>

            <strong className="d-block text-danger my-3">
                Removing teams is irreversible and will cascade to existing child teams.
            </strong>

            {error && <ErrorAlert error={error} />}

            <div className="d-flex justify-content-end pt-1">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    disabled={loading}
                    onClick={onDelete}
                    variant="danger"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Delete team"
                />
            </div>
        </Modal>
    )
}
