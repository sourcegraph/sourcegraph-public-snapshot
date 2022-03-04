import classNames from 'classnames'
import EmailCheckIcon from 'mdi-react/EmailCheckIcon'
import EmailIcon from 'mdi-react/EmailIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { InvitableCollaborator } from '../../auth/welcome/InviteCollaborators/InviteCollaborators'
import { useInviteEmailToSourcegraph } from '../../auth/welcome/InviteCollaborators/useInviteEmailToSourcegraph'
import { CopyableText } from '../../components/CopyableText'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'

import styles from './CollaboratorsPanel.module.scss'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchCollaborators: (userId: string) => Observable<InvitableCollaborator[]>
}

const emailEnabled = window.context?.emailEnabled ?? false

export const CollaboratorsPanel: React.FunctionComponent<Props> = ({
    className,
    authenticatedUser,
    fetchCollaborators,
}) => {
    const inviteEmailToSourcegraph = useInviteEmailToSourcegraph()
    const collaborators = useObservable(
        useMemo(() => fetchCollaborators(authenticatedUser?.id || ''), [fetchCollaborators, authenticatedUser?.id])
    )
    const filteredCollaborators = useMemo(() => collaborators?.slice(0, 6), [collaborators])

    const [inviteError, setInviteError] = useState<ErrorLike | null>(null)
    const [loadingInvites, setLoadingInvites] = useState<Set<string>>(new Set<string>())
    const [successfulInvites, setSuccessfulInvites] = useState<Set<string>>(new Set<string>())

    useEffect(() => {
        if (!Array.isArray(collaborators)) {
            return
        }
        // When Email is not set up we might find some people to invite but won't show that to the user.
        if (!emailEnabled) {
            return
        }

        const loggerPayload = {
            discovered: collaborators.length,
        }
        eventLogger.log('HomepageInvitationsDiscoveredCollaborators', loggerPayload, loggerPayload)
    }, [collaborators])

    const invitePerson = useCallback(
        async (person: InvitableCollaborator): Promise<void> => {
            if (loadingInvites.has(person.email) || successfulInvites.has(person.email)) {
                return
            }
            setLoadingInvites(set => new Set(set).add(person.email))

            try {
                await inviteEmailToSourcegraph({ variables: { email: person.email } })

                setLoadingInvites(set => {
                    const removed = new Set(set)
                    removed.delete(person.email)
                    return removed
                })
                setSuccessfulInvites(set => new Set(set).add(person.email))

                eventLogger.log('HomepageInvitationsSentEmailInvite')
            } catch (error) {
                setInviteError(error)
            }
        },
        [inviteEmailToSourcegraph, loadingInvites, successfulInvites]
    )

    const loadingDisplay = <LoadingPanelView text="Loading colleagues" />

    const contentDisplay =
        filteredCollaborators?.length === 0 || !emailEnabled ? (
            <CollaboratorsPanelNullState username={authenticatedUser?.username || ''} />
        ) : (
            <div className={classNames('row', 'py-1')}>
                {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}

                {filteredCollaborators?.map((person: InvitableCollaborator) => (
                    <div
                        className={classNames(
                            'd-flex',
                            'align-items-center',
                            'col-lg-6',
                            'mt-1',
                            'mb-1',
                            styles.invitebox
                        )}
                        key={person.email}
                    >
                        <Button
                            variant="icon"
                            key={person.email}
                            disabled={loadingInvites.has(person.email) || successfulInvites.has(person.email)}
                            className={classNames('w-100', styles.button)}
                            onClick={() => invitePerson(person)}
                        >
                            <UserAvatar size={40} className={classNames(styles.avatar, 'mr-3')} user={person} />
                            <div className={styles.content}>
                                <strong className={styles.clipText}>{person.displayName}</strong>
                                <div className={styles.inviteButton}>
                                    {loadingInvites.has(person.email) ? (
                                        <span className=" ml-auto mr-3">
                                            <LoadingSpinner inline={true} className="icon-inline mr-1" />
                                            Inviting...
                                        </span>
                                    ) : successfulInvites.has(person.email) ? (
                                        <span className="text-success ml-auto mr-3">
                                            <EmailCheckIcon className="icon-inline mr-1" />
                                            Invited
                                        </span>
                                    ) : (
                                        <>
                                            <div className={classNames('text-muted', styles.clipText)}>
                                                {person.email}
                                            </div>
                                            <div className={classNames('text-primary', styles.inviteButtonOverlay)}>
                                                <EmailIcon className="icon-inline mr-1" />
                                                Invite to Sourcegraph
                                            </div>
                                        </>
                                    )}
                                </div>
                            </div>
                        </Button>
                    </div>
                ))}
            </div>
        )

    return (
        <PanelContainer
            className={classNames(className, 'h-100')}
            title="Invite your colleagues"
            insideTabPanel={true}
            state={collaborators === undefined ? 'loading' : 'populated'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
        />
    )
}

const CollaboratorsPanelNullState: React.FunctionComponent<{ username: string }> = ({ username }) => {
    const inviteURL = `${window.context.externalURL}/sign-up?invitedBy=${username}`

    useEffect(() => {
        const loggerPayload = {
            // The third type, `config-disabled`, is emitted in <HomePanels />
            type: emailEnabled ? 'email-not-configured' : 'no-collaborators',
        }
        eventLogger.log('HomepageInvitationsViewEmpty', loggerPayload, loggerPayload)
    }, [])

    return (
        <div
            className={classNames(
                'd-flex',
                'align-items-center',
                'flex-column',
                'justify-content-center',
                'col-lg-12',
                'h-100'
            )}
        >
            <div className="text-center">No collaborators found in sampled repositories.</div>
            <div className="text-muted mt-3 text-center">
                You can invite people to Sourcegraph with this direct link:
            </div>
            <CopyableText
                className="mt-3"
                text={inviteURL}
                flex={true}
                size={inviteURL.length}
                onCopy={() => eventLogger.log('HomepageInvitationsCopiedInviteLink')}
            />
        </div>
    )
}
