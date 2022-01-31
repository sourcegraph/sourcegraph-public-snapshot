import React, { useEffect } from 'react'

import { Container, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { InviteMemberModal } from './InviteMemberModal'
import { gql, useQuery } from '@apollo/client'
import { OrganizationMembersResult, OrganizationMembersVariables } from '../../graphql-operations'
import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
// import { AddUserToOrgModal } from './addUserToOrgModal'

interface Props
    extends Pick<OrgAreaPageProps, 'org' | 'onOrganizationUpdate' | 'authenticatedUser' | 'isSourcegraphDotCom'> {}

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
export const OrgPendingInvitesPage: React.FunctionComponent<Props> = ({
    org,
    onOrganizationUpdate,
    authenticatedUser,
    isSourcegraphDotCom,
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('OrgPendingInvites', { orgId: org.id })
    }, [org.id])

    const { data, loading, error } = useQuery<OrganizationMembersResult, OrganizationMembersVariables>(
        ORG_MEMBERS_QUERY,
        {
            variables: { id: org.id },
        }
    )

    const viewerCanAddUserToOrganization = !!authenticatedUser && authenticatedUser.siteAdmin

    return (
        <>
            <div className="org-pendinginvites-page">
                <PageTitle title={`${org.name} pending invites`} />
                <div className="d-flex flex-0 justify-content-between align-items-center mb-3">
                    <PageHeader path={[{ text: 'Pending Invites' }]} headingElement="h2" />
                    <div>
                        {/* {viewerCanAddUserToOrganization && !isSourcegraphDotCom && <AddUserToOrgModal orgName={org.name} orgId={org.id} />} */}
                        {viewerCanAddUserToOrganization && <InviteMemberModal orgName={org.name} orgId={org.id} />}
                    </div>
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
