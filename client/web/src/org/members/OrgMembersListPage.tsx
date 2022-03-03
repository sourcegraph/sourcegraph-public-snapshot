import { useMutation, useQuery } from '@apollo/client'
import classNames from 'classnames'
import ChevronDown from 'mdi-react/ChevronDownIcon'
import CogIcon from 'mdi-react/CogIcon'
import React, { useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { pluralize } from '@sourcegraph/common'
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
    Maybe,
    OrganizationMembersResult,
    OrganizationMembersVariables,
    RemoveUserFromOrganizationResult,
    RemoveUserFromOrganizationVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAreaPageProps } from '../area/OrgArea'

import { AddMemberToOrgModal } from './AddMemberToOrgModal'
import { ORG_MEMBERS_QUERY, ORG_MEMBER_REMOVE_MUTATION } from './gqlQueries'
import { IModalInviteResult, InvitedNotification, InviteMemberModalHandler } from './InviteMemberModal'
import styles from './OrgMembersListPage.module.scss'
import { getPaginatedItems, OrgMemberNotification } from './utils'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'> {}
interface Member {
    id: string
    username: string
    displayName: Maybe<string>
    avatarURL: Maybe<string>
}

interface MembersTypeNode {
    viewerCanAdminister: boolean
    members: {
        totalCount: number
        nodes: Member[]
    }
}

interface MemberItemProps {
    member: Member
    viewerCanAdminister: boolean
    isSelf: boolean
    onlyMember: boolean
    orgId: string
    onMemberRemoved: (username: string) => void
}

const MemberItem: React.FunctionComponent<MemberItemProps> = ({
    member,
    orgId,
    viewerCanAdminister,
    isSelf,
    onlyMember,
    onMemberRemoved,
}) => {
    const [removeUserFromOrganization, { loading, error }] = useMutation<
        RemoveUserFromOrganizationResult,
        RemoveUserFromOrganizationVariables
    >(ORG_MEMBER_REMOVE_MUTATION)

    const onRemoveClick = useCallback(async () => {
        if (window.confirm(isSelf ? 'Leave the organization?' : `Remove the user ${member.username}?`)) {
            await removeUserFromOrganization({ variables: { organization: orgId, user: member.id } })
            onMemberRemoved(member.username)
        }
    }, [isSelf, member.username, removeUserFromOrganization, onMemberRemoved, member.id, orgId])

    return (
        <li data-test-username={member.username}>
            <div className="d-flex align-items-center justify-content-between">
                <div
                    className={classNames(
                        'd-flex align-items-center justify-content-start flex-1',
                        styles.memberDetails
                    )}
                >
                    <div className={styles.avatarContainer}>
                        <UserAvatar
                            size={36}
                            className={styles.avatar}
                            user={member}
                            data-tooltip={member.displayName || member.username}
                        />
                    </div>

                    <div className="d-flex flex-column">
                        <Link to={userURL(member.username)}>
                            <strong>{member.displayName || member.username}</strong>
                        </Link>
                        {member.displayName && (
                            <span className={classNames('text-muted', styles.displayName)}>{member.username}</span>
                        )}
                    </div>
                </div>
                <div className={styles.memberRole}>
                    <span className="text-muted">Admin</span>
                </div>
                <div className={styles.memberActions}>
                    {viewerCanAdminister && (
                        <Menu>
                            <MenuButton
                                size="sm"
                                outline={true}
                                variant="secondary"
                                className={styles.memberMenu}
                                disabled={loading}
                            >
                                <CogIcon size={15} />
                                <span aria-hidden={true}>
                                    <ChevronDown size={15} />
                                </span>
                            </MenuButton>

                            <MenuList position={Position.bottomEnd}>
                                <MenuItem onSelect={onRemoveClick} disabled={onlyMember || loading}>
                                    {isSelf ? 'Leave organization...' : 'Remove from organization...'}
                                </MenuItem>
                            </MenuList>
                        </Menu>
                    )}
                    {error && (
                        <ErrorAlert
                            className="mt-2"
                            error={`Error removing ${member.username}. Please, try refreshing the page.`}
                        />
                    )}
                </div>
            </div>
        </li>
    )
}

const MembersResultHeader: React.FunctionComponent<{ total: number; orgName: string }> = ({ total, orgName }) => (
    <li data-test-membersheader="memberslist-header">
        <div className="d-flex align-items-center justify-content-between">
            <div
                className={classNames(
                    'd-flex align-items-center justify-content-start flex-1 member-details',
                    styles.memberDetails
                )}
            >
                {`${total} ${pluralize('person', total, 'people')} in the ${orgName} organization`}
            </div>
            <div className={styles.memberRole}>Role</div>
            <div className={styles.memberActions} />
        </div>
    </li>
)

/**
 * The organization members list page.
 */
export const OrgMembersListPage: React.FunctionComponent<Props> = ({ org, authenticatedUser }) => {
    const [invite, setInvite] = useState<IModalInviteResult>()
    const [notification, setNotification] = useState<string>()
    const [page, setPage] = useState(1)

    const { data, loading, error, refetch } = useQuery<OrganizationMembersResult, OrganizationMembersVariables>(
        ORG_MEMBERS_QUERY,
        {
            variables: { id: org.id },
        }
    )

    useEffect(() => {
        eventLogger.logViewEvent('OrgMembersListV2', { orgId: org.id })
    }, [org.id])

    const isSelf = (userId: string): boolean => authenticatedUser !== null && userId === authenticatedUser.id

    const onInviteSent = useCallback(
        (result: IModalInviteResult) => {
            setInvite(result)
        },
        [setInvite]
    )

    const onInviteSentMessageDismiss = useCallback(() => {
        setInvite(undefined)
    }, [setInvite])

    const onShouldRefetch = useCallback(async () => {
        await refetch({ id: org.id })
    }, [refetch, org.id])

    const onMemberAdded = useCallback(
        async (username: string) => {
            setNotification(`You succesfully added ${username} to ${org.name}`)
            await onShouldRefetch()
        },
        [setNotification, onShouldRefetch, org.name]
    )

    const onMemberRemoved = useCallback(
        async (username: string) => {
            setNotification(`${username} has been removed from the ${org.name} organization on Sourcegraph`)
            setPage(1)
            await onShouldRefetch()
        },
        [setNotification, onShouldRefetch, org.name]
    )

    const onNotificationDismiss = useCallback(() => {
        setNotification(undefined)
    }, [setNotification])

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin
    const membersResult = data ? (data.node as MembersTypeNode) : undefined
    const pagedData = getPaginatedItems(page, membersResult?.members.nodes)

    return (
        <>
            <div className="org-members-page">
                <PageTitle title={`${org.name} Members`} />
                {invite && (
                    <InvitedNotification
                        orgName={org.name}
                        username={invite.username}
                        onDismiss={onInviteSentMessageDismiss}
                        invitationURL={invite.inviteResult.inviteUserToOrganization.invitationURL}
                    />
                )}
                {notification && <OrgMemberNotification message={notification} onDismiss={onNotificationDismiss} />}
                <div className="d-flex flex-0 justify-content-end align-items-center mb-3 flex-wrap">
                    <PageHeader path={[{ text: 'Members' }]} headingElement="h2" className={styles.membersListHeader} />

                    {viewerCanAddUserToOrganization && (
                        <AddMemberToOrgModal orgName={org.name} orgId={org.id} onMemberAdded={onMemberAdded} />
                    )}
                    {authenticatedUser && (
                        <InviteMemberModalHandler
                            variant="success"
                            orgName={org.name}
                            orgId={org.id}
                            onInviteSent={onInviteSent}
                        />
                    )}
                </div>

                <Container className={styles.membersList}>
                    {loading && <LoadingSpinner />}
                    {membersResult && (
                        <ul>
                            <MembersResultHeader total={membersResult.members.totalCount} orgName={org.name} />
                            {pagedData.results.map(usr => (
                                <MemberItem
                                    key={usr.id}
                                    member={usr}
                                    orgId={org.id}
                                    onlyMember={membersResult.members.totalCount === 1}
                                    viewerCanAdminister={membersResult.viewerCanAdminister}
                                    isSelf={isSelf(usr.id)}
                                    onMemberRemoved={onMemberRemoved}
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

                {authenticatedUser &&
                    membersResult &&
                    membersResult.members.totalCount === 1 &&
                    isSelf(membersResult.members.nodes[0].id) && (
                        <Container className={styles.onlyYouContainer}>
                            <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                                <h3>Looks like it’s just you!</h3>
                                <div>
                                    <InviteMemberModalHandler
                                        orgName={org.name}
                                        triggerLabel="Invite a teammate"
                                        orgId={org.id}
                                        onInviteSent={onInviteSent}
                                        className={styles.inviteMemberLink}
                                        as="a"
                                        size="lg"
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
