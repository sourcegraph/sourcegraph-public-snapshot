import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import { Maybe, OrganizationInvitationResponseType } from '@sourcegraph/shared/src/graphql-operations'
import { IEmptyResponse, IOrganizationInvitation } from '@sourcegraph/shared/src/schema'
import { Alert, Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { orgURL } from '..'
import { ModalPage } from '../../components/ModalPage'
import { PageTitle } from '../../components/PageTitle'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAvatar } from '../OrgAvatar'
import { Props as OrgsAreaProps } from '../OrgsArea'

interface Props extends OrgsAreaProps {}

interface RespondToOrgInvitationResult {
    respondToOrganizationInvitation: Maybe<IEmptyResponse>
}

interface RespondToOrgInvitationVariables {
    id: string
    response: OrganizationInvitationResponseType
}

export const RESPOND_TO_ORG_INVITATION = gql`
    mutation RespondToOrgInvitation($id: ID!, $response: OrganizationInvitationResponseType!) {
        respondToOrganizationInvitation(organizationInvitation: $id, responseType: $response) {
            alwaysNil
        }
    }
`

interface InviteResult {
    invitationByJWT: Maybe<IOrganizationInvitation>
}

interface InviteVariables {
    token: string
}

export const INVITATION_BY_JWT = gql`
    query InvitationByJWT($token: String!) {
        invitationByJWT(token: $token) {
            id
            createdAt
            sender {
                username
                displayName
                avatarURL
            }
            organization {
                name
                displayName
            }
        }
    }
`

/**
 * Displays the organization invitation for the user, based on the token in the invite URL.
 */
export const OrgInvitationPage: React.FunctionComponent<Props> = ({ authenticatedUser, history, match }) => {
    const token = match.params.token as string

    const { data: inviteData, loading: inviteLoading, error: inviteError } = useQuery<InviteResult, InviteVariables>(
        INVITATION_BY_JWT,
        {
            skip: !authenticatedUser || !token,
            variables: {
                token,
            },
        }
    )

    const data = inviteData?.invitationByJWT
    const orgName = data?.organization.name
    const sender = data?.sender

    const [respondToInvitation, { loading: respondLoading, error: respondError }] = useMutation<
        RespondToOrgInvitationResult,
        RespondToOrgInvitationVariables
    >(RESPOND_TO_ORG_INVITATION, {
        onError: apolloError => {
            console.error('Error when responding to invitation', apolloError)
        },
    })

    const acceptInvitation = useCallback(
        () =>
            respondToInvitation({
                variables: {
                    id: data?.id || '',
                    response: OrganizationInvitationResponseType.ACCEPT,
                },
            }).then(d => {
                if (orgName) {
                    history.push(orgURL(orgName))
                }
            }),
        [data?.id, history, orgName, respondToInvitation]
    )

    const declineInvitation = useCallback(
        () =>
            respondToInvitation({
                variables: {
                    id: data?.id || '',
                    response: OrganizationInvitationResponseType.REJECT,
                },
            }).then(() => history.push(userURL(authenticatedUser.username))),
        [authenticatedUser.username, data?.id, history, respondToInvitation]
    )

    const loading = inviteLoading || respondLoading
    const error = inviteError || respondError

    return (
        <>
            <PageTitle title={`Invitation to Organization - ${'Hello'}`} />
            {orgName && sender && (
                <ModalPage icon={<OrgAvatar org={orgName} className="mt-3 mb-4" size="lg" />}>
                    <Form className="text-center">
                        <h3>
                            You've been invited to join the {data.organization.displayName || orgName} organization.
                        </h3>
                        <p className="mt-3">
                            <UserAvatar className="mr-2" user={sender} size={24} />
                            <small className="text-muted">
                                Invited by{' '}
                                <Link to={userURL(sender.username)}>{sender.displayName || `@${sender.username}`}</Link>
                                {sender.displayName && <>(@{sender.username})</>}
                            </small>
                        </p>
                        <div className="mt-4 mb-4">
                            <Button className="mr-sm-2" disabled={loading} onClick={acceptInvitation} variant="primary">
                                Join {orgName}
                            </Button>
                            <Button disabled={loading} onClick={declineInvitation} variant="secondary">
                                Decline
                            </Button>
                        </div>
                    </Form>
                </ModalPage>
            )}
            {error && <Alert variant="danger">{error}</Alert>}
            {loading && <LoadingSpinner />}
        </>
    )
}
