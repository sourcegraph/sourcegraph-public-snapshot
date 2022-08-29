import React, { useCallback, useEffect, useState } from 'react'

import { useMutation, useQuery } from '@apollo/client'
import { mdiEmail, mdiCog, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

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
    H3,
    Icon,
    Tooltip,
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
import {
    getInvitationCreationDateString,
    getInvitationExpiryDateString,
    getLocaleFormattedDateFromString,
    getPaginatedItems,
    OrgMemberNotification,
    useQueryStringParameters,
} from './utils'

import styles from './OrgPendingInvites.module.scss'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'> {
    onOrgGetStartedRefresh: () => void
}
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

const PendingInvitesHeader: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
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
    orgId: string
    invite: OrganizationInvitation
    viewerCanAdminister: boolean
    onInviteResentRevoked: (recipient: string, revoked?: boolean) => void
}

const InvitationItem: React.FunctionComponent<React.PropsWithChildren<InvitationItemProps>> = ({
    orgId,
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
        eventLogger.log('OrganizationInviteCopied', { organizationId: orgId }, { organizationId: orgId })
        alert('invite url copied to clipboard!')
    }, [invite.respondURL, orgId])

    const onRevokeInvite = useCallback(async () => {
        eventLogger.log('OrganizationInviteRevokeClicked', { organizationId: orgId }, { organizationId: orgId })

        const inviteToText = invite.recipient ? invite.recipient.username : (invite.recipientEmail as string)
        if (window.confirm(`Revoke invitation from ${inviteToText}?`)) {
            eventLogger.log('OrganizationInviteRevokeConfirmed', { organizationId: orgId }, { organizationId: orgId })
            try {
                await revokeInvite({ variables: { id: invite.id } })
                eventLogger.log('OrgRevokeInvitation', { id: invite.id }, { id: invite.id })
                onInviteResentRevoked(inviteToText, true)
            } catch {
                eventLogger.log('OrgRevokeInvitationError', { id: invite.id }, { id: invite.id })
            }
        } else {
            eventLogger.log('OrganizationInviteRevokeDismissed', { organizationId: orgId }, { organizationId: orgId })
        }
    }, [revokeInvite, onInviteResentRevoked, invite.id, invite.recipientEmail, invite.recipient, orgId])

    const onResendInvite = useCallback(async () => {
        eventLogger.log('OrganizationInviteResendClicked', { organizationId: orgId }, { organizationId: orgId })

        const inviteToText = invite.recipient ? invite.recipient.username : (invite.recipientEmail as string)
        if (window.confirm(`Resend invitation to ${inviteToText}?`)) {
            eventLogger.log('OrganizationInviteResendConfirmed', { organizationId: orgId }, { organizationId: orgId })
            try {
                await resendInvite({ variables: { id: invite.id } })
                eventLogger.log('OrgResendInvitation', { id: invite.id }, { id: invite.id })
                onInviteResentRevoked(inviteToText, false)
            } catch {
                eventLogger.log('OrgResendInvitationError', { id: invite.id }, { id: invite.id })
            }
        } else {
            eventLogger.log('OrganizationInviteResendDismissed', { organizationId: orgId }, { organizationId: orgId })
        }
    }, [orgId, invite.recipient, invite.recipientEmail, invite.id, resendInvite, onInviteResentRevoked])

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
                            <Tooltip content={invite.recipient.displayName || invite.recipient.username}>
                                <UserAvatar size={24} className={styles.avatar} user={invite.recipient} />
                            </Tooltip>
                        )}
                        {!invite.recipient && invite.recipientEmail && (
                            <Icon className={styles.emailIcon} svgPath={mdiEmail} inline={false} aria-hidden={true} />
                        )}
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
                                <Icon svgPath={mdiCog} inline={false} aria-label="Options" height={15} width={15} />
                                <Icon
                                    svgPath={mdiChevronDown}
                                    inline={false}
                                    aria-hidden={true}
                                    height={15}
                                    width={15}
                                />
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
export const OrgPendingInvitesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    org,
    authenticatedUser,
    onOrgGetStartedRefresh,
}) => {
    const orgId = org.id
    const query = useQueryStringParameters()
    const openInviteModal = !!query.get('openInviteModal')
    useEffect(() => {
        eventLogger.logPageView('OrganizationPendingInvites', { organizationId: orgId })
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
            onOrgGetStartedRefresh()
            setInvite(result)
            await refetch({ id: orgId })
        },
        [setInvite, orgId, refetch, onOrgGetStartedRefresh]
    )

    const onInviteResentRevoked = useCallback(
        async (recipient: string, revoked?: boolean) => {
            onOrgGetStartedRefresh()
            const message = `${revoked ? 'Revoked' : 'Resent'} invite for ${recipient}`
            setNotification(message)
            setPage(1)
            await refetch({ id: orgId })
        },
        [setNotification, orgId, refetch, onOrgGetStartedRefresh]
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
                                eventLoggerEventName="InviteMemberButtonClicked"
                                variant="success"
                                initiallyOpened={openInviteModal}
                            />
                        )}
                    </div>
                </div>

                <Container
                    className={classNames(
                        'mb-3',
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
                                    orgId={org.id}
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
                            <H3>No invites pending</H3>
                            <div>
                                <InviteMemberModalHandler
                                    orgName={org.name}
                                    triggerLabel="Invite a teammate"
                                    orgId={org.id}
                                    onInviteSent={onInviteSent}
                                    eventLoggerEventName="InviteMemberCTAClicked"
                                    className={styles.inviteMemberLink}
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
