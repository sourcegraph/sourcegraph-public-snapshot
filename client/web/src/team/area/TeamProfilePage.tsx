import { useCallback, useState } from 'react'

import { mdiPencil } from '@mdi/js'

import { logger } from '@sourcegraph/common'
import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import {
    Button,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxPopover,
    ErrorAlert,
    Form,
    H3,
    Icon,
    Input,
    Label,
    Link,
    Modal,
    Text,
} from '@sourcegraph/wildcard'

import { TEAM_DISPLAY_NAME_MAX_LENGTH } from '..'
import { LoaderButton } from '../../components/LoaderButton'
import { Page } from '../../components/Page'
import { Scalars, TeamAreaTeamFields } from '../../graphql-operations'
import { useTeams } from '../list/backend'

import { useChangeTeamDisplayName, useChangeTeamParent } from './backend'
import { TeamHeader } from './TeamHeader'

import styles from './TeamProfilePage.module.scss'

export interface TeamProfilePageProps {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields

    /** Called when the team is updated and must be reloaded. */
    onTeamUpdate: () => void
}

export const TeamProfilePage: React.FunctionComponent<TeamProfilePageProps> = ({ team, onTeamUpdate }) => {
    const [openModal, setOpenModal] = useState<'edit-display-name' | 'edit-parent' | undefined>()

    const onEditDisplayName = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('edit-display-name')
    }, [])
    const onEditParent = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('edit-parent')
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
                    <H3>Parent team</H3>
                    <Text className="d-flex align-items-center">
                        {team.parentTeam && <span>{team.parentTeam.name}</span>}
                        {!team.parentTeam && <span className="text-muted">No parent team</span>}
                        {team.viewerCanAdminister && (
                            <Button variant="link" onClick={onEditParent} className="ml-2" size="sm">
                                <Icon inline={true} aria-label="Edit team display name" svgPath={mdiPencil} />
                            </Button>
                        )}
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

            {openModal === 'edit-parent' && (
                <EditTeamParentModal
                    onCancel={closeModal}
                    afterEdit={afterAction}
                    teamId={team.id}
                    teamName={team.name}
                    parentName={team.parentTeam?.name ?? null}
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

interface EditTeamParentModalProps {
    teamId: Scalars['ID']
    teamName: string
    parentName: string | null
    onCancel: () => void
    afterEdit: () => void
}

const EditTeamParentModal: React.FunctionComponent<React.PropsWithChildren<EditTeamParentModalProps>> = ({
    teamId,
    teamName,
    parentName: currentParentName,
    onCancel,
    afterEdit,
}) => {
    const labelId = 'editParentTeam'

    const [parentName, setParentName] = useState<string | null>(currentParentName)
    const onParentNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        if (event.currentTarget.value === '') {
            setParentName(null)
        }
        setParentName(event.currentTarget.value)
    }

    const [editTeam, { loading, error }] = useChangeTeamParent()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            if (!event.currentTarget.checkValidity()) {
                return
            }
            try {
                await editTeam({
                    variables: {
                        id: teamId,
                        parentName: parentName ?? '',
                    },
                })
                afterEdit()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [afterEdit, teamId, parentName, editTeam]
    )

    const suggestedTeams = useTeams(parentName)

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Modify parent team of {teamName}</H3>
            {error && <ErrorAlert error={error} />}
            <Form onSubmit={onSubmit}>
                <Combobox aria-label="Choose a repo" style={{ maxWidth: '20rem' }}>
                    <ComboboxInput
                        label="Parent team name"
                        placeholder="parent-team"
                        maxLength={TEAM_DISPLAY_NAME_MAX_LENGTH}
                        autoCorrect="off"
                        autocomplete={false}
                        value={parentName ?? ''}
                        onChange={onParentNameChange}
                        disabled={loading}
                        message="You need to specify repo name (github.com/sg/sg) and then pick one of the suggestions items."
                    />
                    <ComboboxPopover>
                        <ComboboxList>
                            {(suggestedTeams?.connection?.nodes ?? []).map(node => (
                                <ComboboxOption key={node.id} value={node.name} className={styles.item}>
                                    <small className={styles.itemName}>{node.name}</small>
                                    <small className={styles.itemDescription}>{node.displayName}</small>
                                </ComboboxOption>
                            ))}
                        </ComboboxList>
                    </ComboboxPopover>
                </Combobox>
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
