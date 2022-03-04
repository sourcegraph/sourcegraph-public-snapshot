import { useMutation, useQuery } from '@apollo/client'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ChevronDown from 'mdi-react/ChevronDownIcon'
import CogIcon from 'mdi-react/CogIcon'
import EmailIcon from 'mdi-react/EmailIcon'
import React, { useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    Link,
    Menu,
    MenuButton,
    MenuList,
    MenuItem,
    Position,
    PageSelector,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import {
    PendingInvitationsVariables,
    ResendOrgInvitationResult,
    ResendOrgInvitationVariables,
    RevokeInviteResult,
    RevokeInviteVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAreaPageProps } from '../area/OrgArea'

import { ORG_PENDING_INVITES_QUERY, ORG_RESEND_INVITATION_MUTATION, ORG_REVOKE_INVITATION_MUTATION } from './gqlQueries'
import { IModalInviteResult, InvitedNotification, InviteMemberModalHandler } from './InviteMemberModal'
import styles from './OrgPendingInvites.module.scss'
import {
    getInvitationCreationDateString,
    getInvitationExpiryDateString,
    getLocaleFormattedDateFromString,
    getPaginatedItems,
    OrgMemberNotification,
} from './utils'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'> {}
interface OrganizationInvitation {
    id: string
    recipientEmail?: string
    createdAt: string
    notifiedAt: string
    expiresAt: string
    respondURL: string
    recipient?: {
        id: string
        username: string
        displayName: string
        avatarURL: string
    }
    sender: {
        id: string
        displayName?: string
        username: string
    }
    organization: {
        name: string
    }
}

interface IPendingInvitations {
    pendingInvitations: OrganizationInvitation[]
}

const PendingInvitesHeader: React.FunctionComponent = () => (
    <li data-test-pendinginvitesheader="pendingInviiteslist-header">
        <div className="d-flex align-items-center justify-content-between">
            <div
                className={classNames(
                    'd-flex align-items-center justify-content-start flex-1 member-details',
                    styles.inviteDetails
                )}
            >
                Invited member
            </div>
            <div className={styles.inviteDate}>Invited</div>
            <div className={styles.inviteExpiration}>Expiration</div>
            <div className={styles.inviteActions} />
        </div>
    </li>
)

interface InvitationItemProps {
    invite: OrganizationInvitation
    viewerCanAdminister: boolean
    onInviteResentRevoked: (recipient: string, revoked?: boolean) => void
}

const InvitationItem: React.FunctionComponent<InvitationItemProps> = ({
    invite,
    viewerCanAdminister,
    onInviteResentRevoked,
}) => {
    const [revokeInvite, { loading: revokeLoading, error: revokeError }] = useMutation<
        RevokeInviteResult,
        RevokeInviteVariables
    >(ORG_REVOKE_INVITATION_MUTATION)

    const [resendInvite, { loading: resendLoading, error: resendError }] = useMutation<
        ResendOrgInvitationResult,
        ResendOrgInvitationVariables
    >(ORG_RESEND_INVITATION_MUTATION)

    const onCopyInviteLink = useCallback(() => {
        copy(`${window.location.origin}${invite.respondURL}`)
        alert('invite url copied to clipboard!')
    }, [invite.respondURL])

    const onRevokeInvite = useCallback(async () => {
        const inviteToText = invite.recipient ? invite.recipient.username : (invite.recipientEmail as string)
        if (window.confirm(`Revoke invitation from ${inviteToText}?`)) {
            try {
                await revokeInvite({ variables: { id: invite.id } })
                eventLogger.logViewEvent('OrgRevokeInvitation', { id: invite.id })
                onInviteResentRevoked(inviteToText, true)
            } catch {
                eventLogger.logViewEvent('OrgRevokeInvitationError', { id: invite.id })
            }
        }
    }, [revokeInvite, onInviteResentRevoked, invite.id, invite.recipientEmail, invite.recipient])

    const onResendInvite = useCallback(async () => {
        const inviteToText = invite.recipient ? invite.recipient.username : (invite.recipientEmail as string)
        if (window.confirm(`Resend invitation to ${inviteToText}?`)) {
            try {
                await resendInvite({ variables: { id: invite.id } })
                eventLogger.logViewEvent('OrgResendInvitation', { id: invite.id })
                onInviteResentRevoked(inviteToText, false)
            } catch {
                eventLogger.logViewEvent('OrgResendInvitationError', { id: invite.id })
            }
        }
    }, [resendInvite, onInviteResentRevoked, invite.id, invite.recipientEmail, invite.recipient])

    const loading = revokeLoading || resendLoading
    const error = resendError || revokeError

    return (
        <li data-test-username={invite.id}>
            <div className="d-flex align-items-center justify-content-between">
                <div
                    className={classNames(
                        'd-flex align-items-center justify-content-start flex-1',
                        styles.inviteDetails
                    )}
                >
                    <div className={styles.avatarContainer}>
                        {invite.recipient && (
                            <UserAvatar
                                size={24}
                                className={styles.avatar}
                                user={invite.recipient}
                                data-tooltip={invite.recipient.displayName || invite.recipient.username}
                            />
                        )}
                        {!invite.recipient && invite.recipientEmail && <EmailIcon className={styles.emailIcon} />}
                    </div>
                    <div className="d-flex flex-column">
                        {invite.recipient && (
                            <div className="d-flex flex-column">
                                <Link to={userURL(invite.recipient.username)}>
                                    <strong>{invite.recipient.displayName || invite.recipient.username}</strong>
                                </Link>
                                {invite.recipient.displayName && (
                                    <span className={classNames('text-muted', styles.displayName)}>
                                        {invite.recipient.username}
                                    </span>
                                )}
                            </div>
                        )}
                        {invite.recipientEmail && (
                            <span className={styles.recipientEmail}>
                                <strong>{invite.recipientEmail}</strong>
                            </span>
                        )}
                    </div>
                </div>
                <div className={styles.inviteDate}>
                    <span className="text-muted" title={getLocaleFormattedDateFromString(invite.createdAt)}>
                        {`${getInvitationCreationDateString(invite.createdAt)} by ${
                            invite.sender.displayName || invite.sender.username
                        }`}
                    </span>
                </div>
                <div className={styles.inviteExpiration}>
                    <span className="text-muted" title={getLocaleFormattedDateFromString(invite.expiresAt)}>
                        {getInvitationExpiryDateString(invite.expiresAt)}
                    </span>
                </div>
                <div className={styles.inviteActions}>
                    {viewerCanAdminister && (
                        <Menu>
                            <MenuButton
                                size="sm"
                                outline={true}
                                variant="secondary"
                                className={styles.inviteMenu}
                                disabled={loading}
                            >
                                <CogIcon size={15} />
                                <span aria-hidden={true}>
                                    <ChevronDown size={15} />
                                </span>
                            </MenuButton>

                            <MenuList position={Position.bottomEnd}>
                                <MenuItem onSelect={onCopyInviteLink} disabled={loading}>
                                    Copy invite link
                                </MenuItem>
                                <MenuItem onSelect={onResendInvite} disabled={loading}>
                                    Resend invite
                                </MenuItem>
                                <MenuItem onSelect={onRevokeInvite} disabled={loading}>
                                    <span className={styles.revokeInvite}>Revoke invite</span>
                                </MenuItem>
                            </MenuList>
                        </Menu>
                    )}
                    {error && <ErrorAlert className="mt-2" error={error} />}
                </div>
            </div>
        </li>
    )
}

/**
 * The organization members list page.
 */
export const OrgPendingInvitesPage: React.FunctionComponent<Props> = ({ org, authenticatedUser }) => {
    const orgId = org.id
    useEffect(() => {
        eventLogger.logViewEvent('OrgPendingInvites', { orgId })
    }, [orgId])

    const [invite, setInvite] = useState<IModalInviteResult>()
    const [notification, setNotification] = useState<string>()
    const [page, setPage] = useState(1)
    const { data, loading, error, refetch } = useQuery<IPendingInvitations, PendingInvitationsVariables>(
        ORG_PENDING_INVITES_QUERY,
        {
            variables: { id: orgId },
        }
    )

    const onInviteSent = useCallback(
        async (result: IModalInviteResult) => {
            setInvite(result)
            await refetch({ id: orgId })
        },
        [setInvite, orgId, refetch]
    )

    const onInviteResentRevoked = useCallback(
        async (recipient: string, revoked?: boolean) => {
            const message = `${revoked ? 'Revoked' : 'Resent'} invite for ${recipient}`
            setNotification(message)
            setPage(1)
            await refetch({ id: orgId })
        },
        [setNotification, orgId, refetch]
    )

    const onInviteSentMessageDismiss = useCallback(() => {
        setInvite(undefined)
    }, [setInvite])

    const onNotificationDismiss = useCallback(() => {
        setNotification(undefined)
    }, [setNotification])

    const viewerCanInviteUserToOrganization = !!authenticatedUser
    const pagedData = getPaginatedItems(page, data?.pendingInvitations)

    return (
        <>
            <div className="org-pendinginvites-page">
                <PageTitle title={`${org.name} pending invites`} />
                {invite && (
                    <InvitedNotification
                        orgName={org.name}
                        username={invite.username}
                        onDismiss={onInviteSentMessageDismiss}
                        invitationURL={invite.inviteResult.inviteUserToOrganization.invitationURL}
                    />
                )}
                {notification && <OrgMemberNotification message={notification} onDismiss={onNotificationDismiss} />}
                <div className="d-flex flex-0 justify-content-between align-items-center mb-3">
                    <PageHeader path={[{ text: 'Pending Invites' }]} headingElement="h2" />
                    <div>
                        {viewerCanInviteUserToOrganization && (
                            <InviteMemberModalHandler
                                orgName={org.name}
                                orgId={org.id}
                                onInviteSent={onInviteSent}
                                variant="success"
                            />
                        )}
                    </div>
                </div>

                <Container
                    className={classNames(
                        styles.invitationsList,
                        data && !data.pendingInvitations.length && styles.noInvitesList
                    )}
                >
                    {loading && <LoadingSpinner />}
                    {data && (
                        <ul>
                            {data && data.pendingInvitations.length > 0 && <PendingInvitesHeader />}
                            {pagedData.results.map(item => (
                                <InvitationItem
                                    key={item.id}
                                    invite={item}
                                    viewerCanAdminister={true}
                                    onInviteResentRevoked={onInviteResentRevoked}
                                />
                            ))}
                        </ul>
                    )}
                    {error && (
                        <ErrorAlert
                            className="mt-2"
                            error={`Error loading ${org.name} members. Please, try refreshing the page.`}
                        />
                    )}
                </Container>
                {pagedData.totalPages > 1 && (
                    <PageSelector
                        className="mt-4"
                        currentPage={page}
                        onPageChange={setPage}
                        totalPages={pagedData.totalPages}
                    />
                )}
                {authenticatedUser && data && data.pendingInvitations.length === 0 && (
                    <Container>
                        <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                            <h3>No invites pending</h3>
                            <div>
                                <InviteMemberModalHandler
                                    orgName={org.name}
                                    triggerLabel="Invite a teammate"
                                    orgId={org.id}
                                    onInviteSent={onInviteSent}
                                    className={styles.inviteMemberLink}
                                    as="a"
                                    variant="link"
                                />
                                {` to join you on ${org.name} on Sourcegraph`}
                            </div>
                        </div>
                    </Container>
                )}
            </div>
        </>
    )
}
