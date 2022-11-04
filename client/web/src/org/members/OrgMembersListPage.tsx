import React, { useCallback, useEffect, useState } from 'react'

import { useMutation, useQuery } from '@apollo/client'
import { mdiChevronDown, mdiCog } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

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
    H3,
    Icon,
    Tooltip,
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
import { getPaginatedItems, OrgMemberNotification } from './utils'

import styles from './OrgMembersListPage.module.scss'

interface Props
    extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'>,
        RouteComponentProps {
    onOrgGetStartedRefresh: () => void
}
export interface Member {
    id: string
    username: string
    displayName: Maybe<string>
    avatarURL: Maybe<string>
}

export interface MembersTypeNode {
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
    onMemberRemoved: (username: string, isSelf: boolean) => void
}

const MemberItem: React.FunctionComponent<React.PropsWithChildren<MemberItemProps>> = ({
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
        eventLogger.log('RemoveFromOrganizationClicked', { organizationId: orgId }, { organizationId: orgId })
        if (window.confirm(isSelf ? 'Leave the organization?' : `Remove the user ${member.username}?`)) {
            eventLogger.log('RemoveFromOrganizationConfirmed', { organizationId: orgId }, { organizationId: orgId })
            await removeUserFromOrganization({ variables: { organization: orgId, user: member.id } })
            onMemberRemoved(member.username, isSelf)
        } else {
            eventLogger.log('RemoveFromOrganizationDismissed', { organizationId: orgId }, { organizationId: orgId })
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
                        <Tooltip content={member.displayName || member.username}>
                            <UserAvatar size={36} className={styles.avatar} user={member} />
                        </Tooltip>
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
                                <Icon svgPath={mdiCog} inline={false} aria-label="Options" height={15} width={15} />
                                <Icon
                                    svgPath={mdiChevronDown}
                                    inline={false}
                                    height={15}
                                    width={15}
                                    aria-hidden={true}
                                />
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

const MembersResultHeader: React.FunctionComponent<React.PropsWithChildren<{ total: number; orgName: string }>> = ({
    total,
    orgName,
}) => (
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
export const OrgMembersListPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    org,
    authenticatedUser,
    onOrgGetStartedRefresh,
    history,
}) => {
    const [invite, setInvite] = useState<IModalInviteResult>()
    const [notification, setNotification] = useState<string>()
    const [page, setPage] = useState(1)
    const setPageWithEventLogging = useCallback(
        (index: number) => {
            setPage(index)
            eventLogger.log('MemberListPaginationClicked', { organizationId: org.id }, { organizationId: org.id })
        },
        [setPage, org.id]
    )

    const { data, loading, error, refetch } = useQuery<OrganizationMembersResult, OrganizationMembersVariables>(
        ORG_MEMBERS_QUERY,
        {
            variables: { id: org.id },
        }
    )

    useEffect(() => {
        eventLogger.logPageView('OrganizationMembers', { organizationId: org.id })
    }, [org.id])

    const isSelf = (userId: string): boolean => authenticatedUser !== null && userId === authenticatedUser.id

    const onInviteSent = useCallback(
        (result: IModalInviteResult) => {
            setInvite(result)
            onOrgGetStartedRefresh()
        },
        [setInvite, onOrgGetStartedRefresh]
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
            onOrgGetStartedRefresh()
        },
        [setNotification, onShouldRefetch, org.name, onOrgGetStartedRefresh]
    )

    const onMemberRemoved = useCallback(
        async (username: string, isSelf: boolean) => {
            if (isSelf) {
                history.push(userURL(authenticatedUser.username))
            } else {
                setNotification(`${username} has been removed from the ${org.name} organization on Sourcegraph`)
                setPage(1)
                await onShouldRefetch()
                onOrgGetStartedRefresh()
            }
        },
        [org.name, onShouldRefetch, onOrgGetStartedRefresh, history, authenticatedUser.username]
    )

    const onNotificationDismiss = useCallback(() => {
        setNotification(undefined)
    }, [setNotification])

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin
    const membersResult = data ? (data.node as MembersTypeNode) : undefined
    const pagedData = getPaginatedItems(page, membersResult?.members.nodes)
    const showOnlyYou = membersResult && membersResult.members.totalCount === 1

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
                            eventLoggerEventName="InviteMemberButtonClicked"
                        />
                    )}
                </div>

                <Container className={classNames({ 'mb-3': !showOnlyYou }, styles.membersList)}>
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
                        className="mt-4 mb-4"
                        currentPage={page}
                        onPageChange={setPageWithEventLogging}
                        totalPages={pagedData.totalPages}
                    />
                )}

                {authenticatedUser && membersResult && showOnlyYou && isSelf(membersResult.members.nodes[0].id) && (
                    <Container className={styles.onlyYouContainer}>
                        <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                            <H3>Looks like it’s just you!</H3>
                            <div>
                                <InviteMemberModalHandler
                                    orgName={org.name}
                                    triggerLabel="Invite a teammate"
                                    eventLoggerEventName="InviteMemberCTAClicked"
                                    orgId={org.id}
                                    onInviteSent={onInviteSent}
                                    className={styles.inviteMemberLink}
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
