import React, { useCallback, useEffect, useState } from 'react'

import { Container, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { IModalInviteResult, InvitedNotification, InviteMemberModal } from './InviteMemberModal'
import { gql, useQuery } from '@apollo/client'
import { OrganizationMembersResult, OrganizationMembersVariables } from '../../graphql-operations'
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

/**
 * The organization members list page.
 */
export const OrgMembersListPage: React.FunctionComponent<Props> = ({ org, authenticatedUser, isSourcegraphDotCom }) => {
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

                    {viewerCanAddUserToOrganization && isSourcegraphDotCom && (
                        <AddMemberToOrgModal orgName={org.name} orgId={org.id} onMemberAdded={onMemberAdded} />
                    )}
                    {viewerCanAddUserToOrganization && (
                        <InviteMemberModal orgName={org.name} orgId={org.id} onInviteSent={onInviteSent} />
                    )}
                </div>

                <Container>
                    {loading && <LoadingSpinner />}
                    {data && <pre>{JSON.stringify(data, null, 2)}</pre>}
                    {error && (
                        <ErrorAlert
                            className="mt-2"
                            error={`Error loading ${org.name} members. Please, try refreshing the page.`}
                        />
                    )}
                </Container>
            </div>
        </>
    )
}
