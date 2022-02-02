import React, { useCallback, useEffect, useState } from 'react'

import { Container, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { IModalInviteResult, InvitedNotification, InviteMemberModalHandler } from './InviteMemberModal'
import { gql, useQuery } from '@apollo/client'
import { Maybe, OrganizationMembersResult, OrganizationMembersVariables } from '../../graphql-operations'
import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { AddMemberNotification, AddMemberToOrgModal } from './AddMemberToOrgModal'
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
type MembersTypeNode = {
    __typename?: 'Org'
    viewerCanAdminister: boolean
    members: {
        __typename?: 'UserConnection'
        totalCount: number
        nodes: Array<{
            __typename?: 'User'
            id: string
            username: string
            displayName: Maybe<string>
            avatarURL: Maybe<string>
        }>
    }
}
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

    const isSelf = (userId: string): boolean => {
        return authenticatedUser !== null && userId === authenticatedUser.id
    }

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
        [setMemberAdded]
    )

    const onMemberAddedtMessageDismiss = useCallback(() => {
        setMemberAdded(undefined)
    }, [setMemberAdded])

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin

    const members = data ? (data.node as MembersTypeNode).members.nodes : undefined

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

                <Container>
                    {loading && <LoadingSpinner />}
                    {members && <pre>{JSON.stringify(members, null, 2)}</pre>}
                    {error && (
                        <ErrorAlert
                            className="mt-2"
                            error={`Error loading ${org.name} members. Please, try refreshing the page.`}
                        />
                    )}
                </Container>

                {viewerCanAddUserToOrganization && members && members.length === 1 && isSelf(members[0].id) && (
                    <Container className={styles.onlyYouContainer}>
                        <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                            <h3>{"Look like it's just you!"}</h3>
                            <div>
                                <InviteMemberModalHandler
                                    orgName={org.name}
                                    triggerLabel="Invite a teammate"
                                    orgId={org.id}
                                    onInviteSent={onInviteSent}
                                    className={styles.inviteMemberLink}
                                    as="a"
                                    variant="link"
                                ></InviteMemberModalHandler>
                                {` to join you on ${org.name} on Sourcegraph`}
                            </div>
                        </div>
                    </Container>
                )}
            </div>
        </>
    )
}
