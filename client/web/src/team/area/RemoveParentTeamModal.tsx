import React, { useCallback } from 'react'

import { logger } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ErrorAlert, Form, H3, Modal, Text } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type { Scalars } from '../../graphql-operations'

import { useRemoveParentTeam } from './backend'

interface RemoveParentTeamModalProps extends TelemetryV2Props {
    teamID: Scalars['ID']
    teamName: string
    onCancel: () => void
    afterEdit: () => void
}

export const RemoveParentTeamModal: React.FunctionComponent<React.PropsWithChildren<RemoveParentTeamModalProps>> = ({
    teamID,
    teamName,
    onCancel,
    afterEdit,
    telemetryRecorder,
}) => {
    const labelId = 'removeParentTeam'

    const [editTeam, { loading, error }] = useRemoveParentTeam()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            if (!event.currentTarget.checkValidity()) {
                return
            }
            try {
                await editTeam({ variables: { id: teamID } })
                telemetryRecorder.recordEvent('team.parentTeam', 'remove')
                afterEdit()
            } catch (error) {
                telemetryRecorder.recordEvent('team.parentTeam', 'removeFail')
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterEdit, teamID, editTeam, telemetryRecorder]
    )

    return (
        <Modal aria-labelledby={labelId} onDismiss={onCancel}>
            <H3 id={labelId}>Confirm detaching from parent team</H3>
            <Text>
                This change will make {teamName} top-level. That is {teamName} will now have no parent team.
            </Text>
            {error && <ErrorAlert error={error} />}
            <Form onSubmit={onSubmit}>
                <div className="d-flex justify-content-end">
                    <Button className="mr-2" disabled={loading} onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton type="submit" variant="danger" alwaysShowLabel={true} label="Confirm" />
                </div>
            </Form>
        </Modal>
    )
}
