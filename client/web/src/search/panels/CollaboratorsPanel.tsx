import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiEmailCheck, mdiEmail, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Card, CardBody, Link, LoadingSpinner, Icon, H2, Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CopyableText } from '../../components/CopyableText'
import { CollaboratorsFragment, Maybe } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'

import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'
import { useInviteEmailToSourcegraph } from './useInviteEmailToSourcegraph'

import styles from './CollaboratorsPanel.module.scss'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    collaboratorsFragment: CollaboratorsFragment | null
}

const emailEnabled = window.context?.emailEnabled ?? false

export interface InvitableCollaborator {
    email: string
    displayName: string
    name: string
    avatarURL: Maybe<string>
}

export const CollaboratorsPanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    authenticatedUser,
    collaboratorsFragment,
}) => {
    const inviteEmailToSourcegraph = useInviteEmailToSourcegraph()
    const collaborators = collaboratorsFragment?.collaborators ?? null
    const filteredCollaborators = useMemo(() => collaborators?.slice(0, 6), [collaborators])

    const [inviteError, setInviteError] = useState<ErrorLike | null>(null)
    const [loadingInvites, setLoadingInvites] = useState<Set<string>>(new Set<string>())
    const [successfulInvites, setSuccessfulInvites] = useState<Set<string>>(new Set<string>())

    const isSiteAdmin = authenticatedUser?.siteAdmin ?? false

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
            <CollaboratorsPanelNullState username={authenticatedUser?.username || ''} isSiteAdmin={isSiteAdmin} />
        ) : (
            <div className={classNames('row', 'pt-1')}>
                {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}

                <CollaboratorsPanelInfo isSiteAdmin={isSiteAdmin} />

                {filteredCollaborators?.map((person: InvitableCollaborator) => (
                    <div
                        className={classNames('d-flex', 'align-items-center', 'col-lg-6', 'mt-1', styles.invitebox)}
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
                                            <LoadingSpinner className="mr-1" />
                                            Inviting...
                                        </span>
                                    ) : successfulInvites.has(person.email) ? (
                                        <span className="text-success ml-auto mr-3">
                                            <Icon aria-hidden={true} className="mr-1" svgPath={mdiEmailCheck} />
                                            Invited
                                        </span>
                                    ) : (
                                        <>
                                            <div className={classNames('text-muted', styles.clipText)}>
                                                {person.email}
                                            </div>
                                            <div className={classNames('text-primary', styles.inviteButtonOverlay)}>
                                                <Icon aria-hidden={true} className="mr-1" svgPath={mdiEmail} />
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
            className={classNames(className, styles.panel)}
            title="Invite your colleagues"
            insideTabPanel={true}
            state={collaborators === null ? 'loading' : 'populated'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
        />
    )
}

const CollaboratorsPanelNullState: React.FunctionComponent<
    React.PropsWithChildren<{ username: string; isSiteAdmin: boolean }>
> = ({ username, isSiteAdmin }) => {
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
            {emailEnabled ? (
                <div className="text-muted text-center">No collaborators found in sampled repositories.</div>
            ) : isSiteAdmin ? (
                <div className="text-muted text-center">
                    This server is not configured to send emails. <Link to="/help/admin/config/email">Learn more</Link>
                </div>
            ) : null}
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

const CollaboratorsPanelInfo: React.FunctionComponent<React.PropsWithChildren<{ isSiteAdmin: boolean }>> = ({
    isSiteAdmin,
}) => {
    const [infoShown, setInfoShown] = useState<boolean>(false)

    if (infoShown) {
        return (
            <div className="col-12 mt-2 mb-2 position-relative">
                <Card>
                    <CardBody>
                        <div className={classNames('d-flex', 'align-content-start', 'mb-2')}>
                            <H2 className={classNames(styles.infoBox, 'mb-0')}>
                                <Icon aria-hidden={true} className="mr-2 text-muted" svgPath={mdiInformationOutline} />
                                What is this?
                            </H2>
                            <div className="flex-grow-1" />
                            <Button
                                variant="icon"
                                onClick={() => setInfoShown(false)}
                                aria-label="Close info box"
                                aria-expanded="true"
                            >
                                <span aria-hidden="true">×</span>
                            </Button>
                        </div>
                        {isSiteAdmin ? (
                            <>
                                <Text className={styles.infoBox}>
                                    This feature enables Sourcegraph users to invite collaborators we discover through
                                    your Git repository commit history. The invitee will receive a link to Sourcegraph,
                                    but no special permissions are granted.
                                </Text>
                                <Text className={classNames(styles.infoBox, 'mb-0')}>
                                    If you wish to disable this feature, see{' '}
                                    <Link to="/help/admin/config/user_invitations">this documentation</Link>.
                                </Text>
                            </>
                        ) : (
                            <Text className={classNames(styles.infoBox, 'mb-0')}>
                                These collaborators were found via your repositories Git commit history. The invitee
                                will receive a link to Sourcegraph, but no special permissions are granted.
                            </Text>
                        )}
                    </CardBody>
                </Card>
            </div>
        )
    }
    return (
        <div className={classNames('col-12', 'd-flex', 'mt-2', 'mb-1')}>
            <div className={classNames('text-muted', styles.info)}>Collaborators from your repositories</div>
            <div className="flex-grow-1" />
            <div>
                <Icon aria-hidden={true} className="mr-1 text-muted" svgPath={mdiInformationOutline} />
                <Button
                    variant="link"
                    className={classNames(styles.info, 'p-0')}
                    onClick={() => setInfoShown(true)}
                    aria-haspopup="true"
                    aria-expanded="false"
                >
                    What is this?
                </Button>
            </div>
        </div>
    )
}
