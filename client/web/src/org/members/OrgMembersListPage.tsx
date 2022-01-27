import React, { useCallback, useEffect } from 'react'

import { Container, PageHeader, Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { InviteMemberModal } from './InviteMemberModal'
import { gql, useQuery } from '@apollo/client'
import { OrganizationMembersResult, OrganizationMembersVariables } from '../../graphql-operations'
import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'onOrganizationUpdate'> {}

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
export const OrgMembersListPage: React.FunctionComponent<Props> = ({ org, onOrganizationUpdate }) => {
    useEffect(() => {
        eventLogger.logViewEvent('OrgSettingsProfile')
    }, [org.id])

    const { data, loading, error } = useQuery<OrganizationMembersResult, OrganizationMembersVariables>(
        ORG_MEMBERS_QUERY,
        {
            variables: { id: org.id },
        }
    )

    const [modalInvite, setModalInvite] = React.useState(false)

    const onInviteClick = useCallback(() => {
        setModalInvite(true)
    }, [setModalInvite])

    const onCloseIviteModal = useCallback(() => {
        setModalInvite(false)
    }, [setModalInvite])

    return (
        <>
            <div className="org-members-page">
                <PageTitle title={`${org.name} Members`} />
                <div className="d-flex flex-0 justify-content-between align-items-start">
                    <PageHeader path={[{ text: 'Organization Members' }]} headingElement="h2" className="mb-3" />
                    <Button variant="success" onClick={onInviteClick}>
                        Invite member
                    </Button>
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
            {modalInvite && <InviteMemberModal onClose={onCloseIviteModal} orgName={org.name} />}
        </>
    )
}
