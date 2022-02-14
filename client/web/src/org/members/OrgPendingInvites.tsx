import { gql, useQuery } from '@apollo/client'
// import { MenuItem, MenuList } from '@reach/menu-button'
import classNames from 'classnames'
import CogIcon from 'mdi-react/CogIcon'
import EmailIcon from 'mdi-react/EmailIcon'
import React, { useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Container, PageHeader, LoadingSpinner, Link, Menu, MenuButton } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { PendingInvitationsVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAreaPageProps } from '../area/OrgArea'

import { IModalInviteResult, InvitedNotification, InviteMemberModalHandler } from './InviteMemberModal'
import styles from './OrgPendingInvites.module.scss'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'> {}

const ORG_PENDING_INVITES_QUERY = gql`
    query PendingInvitations($id: ID!) {
    pendingInvitations(organization: $id) {
        id
        recipientEmail
        recipient {
        id
        username
        displayName
        }
        revokedAt
        sender {
        id
        displayName
        username
        }
        organization {
        name
        }
        createdAt
        notifiedAt
    }
    }
`

interface OrganizationInvitation {
    id: string;
        recipientEmail?: string;
        revokedAt: string;
        createdAt: string;
        notifiedAt: string;
        recipient?: {
            id: string;
            username: string;
            displayName: string;
            avatarURL: string;
        };
        sender: {
            id: string;
            displayName?: string;
            username: string;
        };
        organization: {
            name: string;
        };
    }

interface IPendingInvitations {
    pendingInvitations: OrganizationInvitation[]
}

interface InvitationItemProps {
    invite: OrganizationInvitation,
    viewerCanAdminister: boolean
    orgId: string
    onShouldRefetch: () => void
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
        <div className={styles.inviteDate}>
                    Invited
                </div>
                <div className={styles.inviteExpiration}>
                    Expiration
                </div>
                <div className={styles.inviteActions} />
    </div>
</li>
)

const InvitationItem: React.FunctionComponent<InvitationItemProps> = ({
    invite,
    orgId,
    viewerCanAdminister,
    onShouldRefetch,
}) =>
    // const [removeUserFromOrganization, { loading, error }] = useMutation<
    //     RemoveUserFromOrganizationResult,
    //     RemoveUserFromOrganizationVariables
    // >(ORG_MEMBER_REMOVE_QUERY)

    // const onRemoveClick = useCallback(async () => {
    //     if (window.confirm(isSelf ? 'Leave the organization?' : `Remove the user ${member.username}?`)) {
    //         await removeUserFromOrganization({ variables: { organization: orgId, user: member.id } })
    //         onShouldRefetch()
    //     }
    // }, [isSelf, member.username, removeUserFromOrganization, onShouldRefetch, member.id, orgId])

  (
        <li data-test-username={invite.id}>
            <div className="d-flex align-items-center justify-content-between">
                <div
                    className={classNames(
                        'd-flex align-items-center justify-content-start flex-1',
                        styles.inviteDetails
                    )}
                >
                    <div className={styles.avatarContainer}>
                    { invite.recipient && <UserAvatar
                            className={styles.avatar}
                            user={invite.recipient}
                            data-tooltip={invite.recipient.displayName || invite.recipient.username}
                        />}
                        {invite.recipientEmail && <EmailIcon className={classNames(styles.avatar, styles.emailIcon)} />}
                    </div>
                    <div className="d-flex flex-column">
                        { invite.recipient && (<div className="d-flex flex-column">
                                <Link to={userURL(invite.recipient.username)}>
                                    <strong>{invite.recipient.displayName || invite.recipient.username}</strong>
                                </Link>
                                {invite.recipient.displayName && <span className="text-muted">{invite.recipient.username}</span>}
                            </div>)}
                            {invite.recipientEmail && <span className={styles.recipientEmail}><strong>{invite.recipientEmail}</strong></span>}
                    </div>
                </div>
                <div className={styles.inviteDate}>
                    <span className="text-muted">Admin</span>
                </div>
                <div className={styles.inviteExpiration}>
                    <span className="text-muted">Admin</span>
                </div>
                <div className={styles.inviteActions}>
                    {viewerCanAdminister && (
                        <Menu>
                            <MenuButton variant="secondary" outline={false} className={styles.inviteMenu}>
                                {/* {loading ? <LoadingSpinner /> : <CogIcon />} */}
                                <span aria-hidden={true}>▾</span>
                            </MenuButton>

                            {/* <MenuList>
                                <MenuItem onSelect={onRemoveClick} disabled={onlyMember || loading}>
                                    {isSelf ? 'Leave organization' : 'Remove from organization'}
                                </MenuItem>
                            </MenuList> */}
                        </Menu>
                    )}
                    {/* {error && (
                        <ErrorAlert
                            className="mt-2"
                            error={`Error removing ${member.username}. Please, try refreshing the page.`}
                        />
                    )} */}
                </div>
            </div>
        </li>
    )

/**
 * The organization members list page.
 */
export const OrgPendingInvitesPage: React.FunctionComponent<Props> = ({ org, authenticatedUser }) => {
    const orgId = org.id
    useEffect(() => {
        eventLogger.logViewEvent('OrgPendingInvites', { orgId })
    }, [orgId])

    const [invite, setInvite] = useState<IModalInviteResult>()
    const { data, loading, error, refetch }= useQuery<IPendingInvitations, PendingInvitationsVariables>(
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

    const onInviteSentMessageDismiss = useCallback(() => {
        setInvite(undefined)
    }, [setInvite])

    const onShouldRefetch = useCallback(
        async () => {
            await refetch({ id: org.id })
        },
        [refetch, org.id]
    )

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin

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
                <div className="d-flex flex-0 justify-content-between align-items-center mb-3">
                    <PageHeader path={[{ text: 'Pending Invites' }]} headingElement="h2" />
                    <div>
                        {viewerCanAddUserToOrganization && (
                            <InviteMemberModalHandler
                                orgName={org.name}
                                orgId={org.id}
                                onInviteSent={onInviteSent}
                                variant="success"
                            />
                        )}
                    </div>
                </div>

                <Container className={styles.invitationsList}>
                    {loading && <LoadingSpinner />}
                    {data && (
                        <ul>
                            <PendingInvitesHeader  />
                            {data.pendingInvitations.map(item => (
                                <InvitationItem
                                    key={item.id}
                                    invite={item}
                                    orgId={org.id}
                                    viewerCanAdminister={true}
                                    onShouldRefetch={onShouldRefetch}
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
                {authenticatedUser &&
                    data &&
                    data.pendingInvitations.length === 0 &&
                        (<Container className={styles.noInvitesContainer}>
                            <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                                <h3>Look like it's just you!</h3>
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
