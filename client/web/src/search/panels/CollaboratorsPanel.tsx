import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { Observable } from 'rxjs'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { InvitableCollaborator } from '../../auth/welcome/InviteCollaborators/InviteCollaborators'
import { UserAvatar } from '../../user/UserAvatar'

import styles from './CollaboratorsPanel.module.scss'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchCollaborators: (userId: string) => Observable<InvitableCollaborator[]>
}

export const CollaboratorsPanel: React.FunctionComponent<Props> = ({
    className,
    authenticatedUser,
    fetchCollaborators,
    telemetryService,
}) => {
    const collaborators = useObservable(
        useMemo(() => fetchCollaborators(authenticatedUser?.id || ''), [fetchCollaborators, authenticatedUser?.id])
    )
    const filteredCollaborators = useMemo(() => collaborators?.slice(0, 4), [collaborators])

    const [inviteError, setInviteError] = useState<ErrorLike | null>(null)
    const [loadingInvites, setLoadingInvites] = useState<Set<string>>(new Set<string>())
    const [successfulInvites, setSuccessfulInvites] = useState<Set<string>>(new Set<string>())
    const invitePerson = useCallback(
        async (person: InvitableCollaborator): Promise<void> => {
            if (loadingInvites.has(person.email) || successfulInvites.has(person.email)) {
                return
            }
            setLoadingInvites(set => new Set(set).add(person.email))

            try {
                // await inviteEmailToSourcegraph({ variables: { email: person.email } })
                await new Promise(resolve => setTimeout(resolve, 2000))

                setLoadingInvites(set => {
                    const removed = new Set(set)
                    removed.delete(person.email)
                    return removed
                })
                setSuccessfulInvites(set => new Set(set).add(person.email))

                // eventLogger.log('UserInvitationsSentEmailInvite')
            } catch (error) {
                setInviteError(error)
            }
        },
        [loadingInvites, successfulInvites]
    )

    const loadingDisplay = <LoadingPanelView text="Loading colleagues" />

    const contentDisplay = (
        <div className={classNames('d-flex', 'flex-row', 'py-4')}>
            {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}

            {filteredCollaborators?.map((person: InvitableCollaborator) => (
                <div className={classNames('d-flex', 'ml-3', 'align-items-center', 'flex-grow-1')} key={person.email}>
                    <UserAvatar size={64} className={classNames(styles.avatar, 'mr-3')} user={person} />
                    <div>
                        <strong>{person.displayName}</strong>
                        <div className="text-muted">{person.email}</div>
                        <div className={styles.inviteButton}>
                            {loadingInvites.has(person.email) ? (
                                <LoadingSpinner inline={true} className={classNames('ml-auto', 'mr-3')} />
                            ) : successfulInvites.has(person.email) ? (
                                <span className="text-muted ml-auto mr-3">
                                    <CheckCircleIcon className="icon-inline mr-1" />
                                    Invited
                                </span>
                            ) : (
                                <Button
                                    variant="secondary"
                                    outline={true}
                                    size="sm"
                                    className={classNames('ml-auto', 'mr-3')}
                                    onClick={() => invitePerson(person)}
                                >
                                    Invite
                                </Button>
                            )}
                        </div>
                    </div>
                </div>
            ))}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'repositories-panel')}
            title="Invite your colleagues"
            state={collaborators === undefined ? 'loading' : 'populated'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
        />
    )
}
