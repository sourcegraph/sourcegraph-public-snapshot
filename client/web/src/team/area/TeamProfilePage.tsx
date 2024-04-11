import React, { useCallback, useEffect, useState } from 'react'

import { mdiDelete, mdiPencil } from '@mdi/js'

import { logger } from '@sourcegraph/common'
import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { type TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ErrorAlert, Form, H3, Icon, Input, Label, Link, Modal, Text } from '@sourcegraph/wildcard'

import { TEAM_DISPLAY_NAME_MAX_LENGTH } from '..'
import { LoaderButton } from '../../components/LoaderButton'
import { Page } from '../../components/Page'
import type { Scalars, TeamAreaTeamFields } from '../../graphql-operations'

import { useChangeTeamDisplayName } from './backend'
import { EditParentTeamModal } from './EditParentTeamModal'
import { RemoveParentTeamModal } from './RemoveParentTeamModal'
import { TeamHeader } from './TeamHeader'

export interface TeamProfilePageProps extends TelemetryV2Props {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields

    /** Called when the team is updated and must be reloaded. */
    onTeamUpdate: () => void
}

export const TeamProfilePage: React.FunctionComponent<TeamProfilePageProps> = ({
    team,
    onTeamUpdate,
    telemetryRecorder,
}) => {
    const [openModal, setOpenModal] = useState<
        'edit-display-name' | 'edit-parent-team' | 'remove-parent-team' | undefined
    >()

    useEffect(() => telemetryRecorder.recordEvent('team.profile', 'view'), [telemetryRecorder])

    const onEditDisplayName = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('edit-display-name')
    }, [])
    const onEditParentTeam = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('edit-parent-team')
    }, [])
    const onConfirmParentTeamRemoval = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('remove-parent-team')
    }, [])
    const closeModal = useCallback(() => {
        setOpenModal(undefined)
    }, [])
    const afterAction = useCallback(() => {
        setOpenModal(undefined)
        onTeamUpdate()
    }, [onTeamUpdate])

    return (
        <>
            <Page className="mb-3">
                <TeamHeader team={team} className="mb-3" />
                <div className="container">
                    <H3>Team name</H3>
                    <Text>
                        <TeamAvatar team={team} inline={true} className="mr-1" />
                        {team.name}
                    </Text>
                    <H3>Display Name</H3>
                    <Text className="d-flex align-items-center">
                        {team.displayName && <span>{team.displayName}</span>}
                        {!team.displayName && <span className="text-muted">No display name set</span>}{' '}
                        {team.viewerCanAdminister && (
                            <Button variant="link" onClick={onEditDisplayName} className="ml-2" size="sm">
                                <Icon inline={true} aria-label="Edit team display name" svgPath={mdiPencil} />
                            </Button>
                        )}
                    </Text>
                    <H3>Parent team</H3>
                    <Text className="d-flex align-items-center">
                        {team.parentTeam && <span>{team.parentTeam?.displayName || team.parentTeam?.name}</span>}
                        {!team.parentTeam && <span className="text-muted">Root team - no parent</span>}{' '}
                        {team.viewerCanAdminister && (
                            <Button variant="link" onClick={onEditParentTeam} className="ml-2" size="sm">
                                <Icon
                                    inline={true}
                                    aria-label={team.parentTeam ? 'Edit parent team' : 'Add parent team'}
                                    svgPath={mdiPencil}
                                />
                            </Button>
                        )}
                        {team.viewerCanAdminister && team.parentTeam && (
                            <Button variant="link" onClick={onConfirmParentTeamRemoval} className="ml-2" size="sm">
                                <Icon inline={true} aria-label="Remove parent team" svgPath={mdiDelete} />
                            </Button>
                        )}
                    </Text>
                    <H3>Creator</H3>
                    <Text className="d-flex align-items-center">
                        {team.creator !== null && (
                            <>
                                <UserAvatar user={team.creator} inline={true} className="mr-1" />
                                <Link to={team.creator.url}>
                                    {team.creator.displayName ? team.creator.displayName : team.creator.username}
                                </Link>
                            </>
                        )}
                        {team.creator === null && <span className="text-muted">Deleted user</span>}
                    </Text>
                </div>
            </Page>

            {openModal === 'edit-display-name' && (
                <EditTeamDisplayNameModal
                    onCancel={closeModal}
                    afterEdit={afterAction}
                    teamID={team.id}
                    teamName={team.name}
                    displayName={team.displayName}
                />
            )}

            {openModal === 'edit-parent-team' && (
                <EditParentTeamModal
                    telemetryRecorder={telemetryRecorder}
                    onCancel={closeModal}
                    afterEdit={afterAction}
                    teamID={team.id}
                    teamName={team.name}
                    parentTeamName={team.parentTeam?.name || null}
                />
            )}

            {openModal === 'remove-parent-team' && (
                <RemoveParentTeamModal
                    telemetryRecorder={telemetryRecorder}
                    onCancel={closeModal}
                    afterEdit={afterAction}
                    teamID={team.id}
                    teamName={team.name}
                />
            )}
        </>
    )
}

interface EditTeamDisplayNameModalProps {
    teamID: Scalars['ID']
    teamName: string
    displayName: string | null

    onCancel: () => void
    afterEdit: () => void
}

const EditTeamDisplayNameModal: React.FunctionComponent<React.PropsWithChildren<EditTeamDisplayNameModalProps>> = ({
    teamID,
    teamName,
    displayName: currentDisplayName,
    onCancel,
    afterEdit,
}) => {
    const labelId = 'editDisplayName'

    const [displayName, setDisplayName] = useState<string>(currentDisplayName ?? '')
    const onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setDisplayName(event.currentTarget.value)
    }

    const [editTeam, { loading, error }] = useChangeTeamDisplayName()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()

            if (!event.currentTarget.checkValidity()) {
                return
            }

            try {
                await editTeam({ variables: { id: teamID, displayName: displayName ?? null } })

                afterEdit()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterEdit, teamID, displayName, editTeam]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Modify team {teamName} display name</H3>

            {error && <ErrorAlert error={error} />}

            <Form onSubmit={onSubmit}>
                <Label htmlFor="edit-team--displayname" className="mt-2">
                    Display name
                </Label>
                <Input
                    id="edit-team--displayname"
                    placeholder="Engineering Team"
                    maxLength={TEAM_DISPLAY_NAME_MAX_LENGTH}
                    autoCorrect="off"
                    value={displayName}
                    onChange={onDisplayNameChange}
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
