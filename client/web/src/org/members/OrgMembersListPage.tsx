import { gql, useQuery } from '@apollo/client'
import { MenuItem, MenuList } from '@reach/menu-button'
import classNames from 'classnames'
import CogIcon from 'mdi-react/CogIcon'
import React, { useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Container, PageHeader, LoadingSpinner, Link, Menu, MenuButton } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { Maybe, OrganizationMembersResult, OrganizationMembersVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAreaPageProps } from '../area/OrgArea'

import { AddMemberNotification, AddMemberToOrgModal } from './AddMemberToOrgModal'
import { IModalInviteResult, InvitedNotification, InviteMemberModalHandler } from './InviteMemberModal'
import styles from './OrgMembersListPage.module.scss'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'authenticatedUser' | 'isSourcegraphDotCom'> {}

const ORG_MEMBERS_QUERY = gql`
    query OrganizationMembers($id: ID!) {
        node(id: $id) {
            ... on Org {
                viewerCanAdminister
                members {
                    nodes {
                        id
                        username
                        displayName
                        avatarURL
                    }
                    totalCount
                }
            }
        }
    }
`
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

const MemberItem: React.FunctionComponent<{ member: Member; viewerCanAdminister: boolean; isSelf: boolean }> = ({
    member,
    viewerCanAdminister,
    isSelf,
}) => (
    <li data-test-username={member.username}>
        <div className="d-flex align-items-center justify-content-between">
            <div
                className={classNames(
                    'd-flex align-items-center justify-content-start flex-1 member-details',
                    styles.memberDetails
                )}
            >
                <div className={styles.avatarContainer}>
                    <UserAvatar
                        className={styles.avatar}
                        user={member}
                        data-tooltip={member.displayName || member.username}
                    />
                </div>
                <div className="d-flex flex-column">
                    <Link to={userURL(member.username)}>
                        <strong>{member.displayName || member.username}</strong>
                    </Link>
                    {member.displayName && <span className="text-muted">{member.username}</span>}
                </div>
            </div>
            <div className="site-admin-detail-list__actions">
                {viewerCanAdminister && (
                    <Menu>
                        <MenuButton variant="secondary" outline={false}>
                            <CogIcon /> <span aria-hidden={true}>▾</span>
                        </MenuButton>

                        <MenuList>
                            <MenuItem onSelect={() => alert('Clicked!')}>
                                {isSelf ? 'Leave organization' : 'Remove from organization'}
                            </MenuItem>
                        </MenuList>
                    </Menu>
                )}
            </div>
        </div>
    </li>
)

/**
 * The organization members list page.
 */
export const OrgMembersListPage: React.FunctionComponent<Props> = ({ org, authenticatedUser }) => {
    const [invite, setInvite] = useState<IModalInviteResult>()
    const [member, setMemberAdded] = useState<string>()

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

    const onMemberAdded = useCallback(
        async (username: string) => {
            setMemberAdded(username)
            await refetch({ id: org.id })
        },
        [setMemberAdded, refetch, org.id]
    )

    const onMemberAddedtMessageDismiss = useCallback(() => {
        setMemberAdded(undefined)
    }, [setMemberAdded])

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin

    const membersResult = data ? (data.node as MembersTypeNode) : undefined

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
                {member && (
                    <AddMemberNotification
                        orgName={org.name}
                        username={member}
                        onDismiss={onMemberAddedtMessageDismiss}
                    />
                )}
                <div className="d-flex flex-0 justify-content-end align-items-center mb-3 flex-wrap">
                    <PageHeader
                        path={[{ text: 'Organization Members' }]}
                        headingElement="h2"
                        className={styles.membersListHeader}
                    />

                    {viewerCanAddUserToOrganization && (
                        <AddMemberToOrgModal orgName={org.name} orgId={org.id} onMemberAdded={onMemberAdded} />
                    )}
                    {viewerCanAddUserToOrganization && (
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
                            <li>{`${membersResult.members.totalCount} ${pluralize(
                                'person',
                                membersResult.members.totalCount
                            )} in the ${org.name} organization`}</li>
                            {membersResult.members.nodes.map(usr => (
                                <MemberItem
                                    key={usr.id}
                                    member={usr}
                                    viewerCanAdminister={membersResult.viewerCanAdminister}
                                    isSelf={isSelf(usr.id)}
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

                {viewerCanAddUserToOrganization &&
                    membersResult &&
                    membersResult.members.totalCount === 1 &&
                    isSelf(membersResult.members.nodes[0].id) && (
                        <Container className={styles.onlyYouContainer}>
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
