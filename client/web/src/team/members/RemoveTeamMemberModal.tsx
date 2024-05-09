import React, { useCallback } from 'react'

import { logger } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, H3, Modal, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type { ListTeamMemberFields, Scalars } from '../../graphql-operations'

import { useRemoveTeamMembers } from './backend'

export interface RemoveTeamMemberModalProps extends TelemetryV2Props {
    teamID: Scalars['ID']
    teamName: string
    member: ListTeamMemberFields

    onCancel: () => void
    afterRemove: () => void
}

export const RemoveTeamMemberModal: React.FunctionComponent<React.PropsWithChildren<RemoveTeamMemberModalProps>> = ({
    teamID,
    teamName,
    member,
    onCancel,
    afterRemove,
    telemetryRecorder,
}) => {
    const labelId = 'removeTeamMember'

    const [removeMembers, { loading, error }] = useRemoveTeamMembers()

    const onRemove = useCallback<React.MouseEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await removeMembers({ variables: { team: teamID, members: [{ userID: member.id }] } })

                telemetryRecorder.recordEvent('team.members', 'remove')
                afterRemove()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
                telemetryRecorder.recordEvent('team.members', 'removeFail')
            }
        },
        [afterRemove, teamID, member.id, removeMembers, telemetryRecorder]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>
                Remove {member.username} from {teamName}?
            </H3>

            {error && <ErrorAlert error={error} />}

            <div className="d-flex justify-content-end pt-1">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    disabled={loading}
                    onClick={onRemove}
                    variant="danger"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Remove member"
                />
            </div>
        </Modal>
    )
}
