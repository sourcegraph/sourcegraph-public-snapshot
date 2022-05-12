import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import { Maybe, OrganizationInvitationResponseType } from '@sourcegraph/shared/src/graphql-operations'
import { IEmptyResponse, IOrganizationInvitation } from '@sourcegraph/shared/src/schema'
import { Alert, AnchorLink, Button, LoadingSpinner, Link, Typography } from '@sourcegraph/wildcard'

import { orgURL } from '..'
import { AuthenticatedUser } from '../../auth'
import { ModalPage } from '../../components/ModalPage'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgAvatar } from '../OrgAvatar'

import styles from './OrgInvitationPage.module.scss'

interface Props extends RouteComponentProps<{ token: string }> {
    authenticatedUser: AuthenticatedUser
    className?: string
}

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
    invitationByToken: Maybe<IOrganizationInvitation>
}

interface InviteVariables {
    token: string
}

export const INVITATION_BY_TOKEN = gql`
    query InvitationByToken($token: String!) {
        invitationByToken(token: $token) {
            createdAt
            id
            isVerifiedEmail
            organization {
                displayName
                name
            }
            recipientEmail
            sender {
                avatarURL
                displayName
                username
            }
        }
    }
`

/**
 * Displays the organization invitation for the user, based on the token in the invite URL.
 */
export const OrgInvitationPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    className,
    history,
    match,
}) => {
    const token = match.params.token

    const { data: inviteData, loading: inviteLoading, error: inviteError } = useQuery<InviteResult, InviteVariables>(
        INVITATION_BY_TOKEN,
        {
            skip: !authenticatedUser || !token,
            variables: {
                token,
            },
        }
    )

    const data = inviteData?.invitationByToken
    const orgName = data?.organization.name
    const orgId = data?.organization.id
    const sender = data?.sender
    const orgDisplayName = data?.organization.displayName || orgName
    const willVerifyEmail = data?.recipientEmail && !data?.isVerifiedEmail

    useEffect(() => {
        eventLogger.logPageView('OrganizationInvitation', { organizationId: orgId, invitationId: data?.id })
    }, [orgId, data?.id])

    const [respondToInvitation, { loading: respondLoading, error: respondError }] = useMutation<
        RespondToOrgInvitationResult,
        RespondToOrgInvitationVariables
    >(RESPOND_TO_ORG_INVITATION, {
        onError: apolloError => {
            console.error('Error when responding to invitation', apolloError)
        },
    })

    const acceptInvitation = useCallback(async () => {
        eventLogger.log(
            'OrganizationInvitationAcceptClicked',
            {
                organizationId: orgId,
                invitationId: data?.id,
                willVerifyEmail,
            },
            {
                organizationId: orgId,
                invitationId: data?.id,
                willVerifyEmail,
            }
        )
        try {
            await respondToInvitation({
                variables: {
                    id: data?.id || '',
                    response: OrganizationInvitationResponseType.ACCEPT,
                },
            })
            eventLogger.log(
                'OrganizationInvitationAcceptSucceeded',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        } catch {
            eventLogger.log(
                'OrganizationInvitationAcceptFailed',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
            return
        }

        if (orgName) {
            history.push(orgURL(orgName))
        }
    }, [data?.id, history, orgId, orgName, respondToInvitation, willVerifyEmail])

    const declineInvitation = useCallback(async () => {
        eventLogger.log(
            'OrganizationInvitationDeclineClicked',
            {
                organizationId: orgId,
                invitationId: data?.id,
                willVerifyEmail,
            },
            {
                organizationId: orgId,
                invitationId: data?.id,
                willVerifyEmail,
            }
        )
        try {
            await respondToInvitation({
                variables: {
                    id: data?.id || '',
                    response: OrganizationInvitationResponseType.REJECT,
                },
            })
            eventLogger.log(
                'OrganizationInvitationDeclineSucceeded',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        } catch {
            eventLogger.log(
                'OrganizationInvitationDeclineFailed',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        }

        history.push(userURL(authenticatedUser.username))
    }, [authenticatedUser.username, data?.id, history, orgId, respondToInvitation, willVerifyEmail])

    const loading = inviteLoading || respondLoading
    const error = inviteError?.message || respondError?.message

    return (
        <>
            <PageTitle title={`Invitation to Organization ${orgName || ''}`} />
            {orgName && sender && (
                <ModalPage
                    className={classNames(styles.orgInvitationPage, className)}
                    icon={<OrgAvatar org={orgName} className="mt-3 mb-4" size="lg" />}
                >
                    <Form className="text-center pr-4 pl-4 pb-4">
                        <Typography.H2>You've been invited to join the {orgDisplayName} organization</Typography.H2>
                        <div className="mt-4">
                            <UserAvatar className={classNames('mr-2', styles.userAvatar)} user={sender} size={24} />
                            <span>
                                Invited by{' '}
                                <Link to={userURL(sender.username)}>{sender.displayName || `@${sender.username}`}</Link>
                                {sender.displayName && <span className="text-muted">(@{sender.username})</span>}
                            </span>
                        </div>
                        {data.isVerifiedEmail === false && data.recipientEmail && (
                            <div className="mt-4 mb-4">
                                This invite was sent to <strong>{data.recipientEmail}</strong>. Joining the{' '}
                                {orgDisplayName} organization will add this as a verified email on your account.
                            </div>
                        )}
                        <div className="mt-4">
                            <Button className="mr-sm-2" disabled={loading} onClick={acceptInvitation} variant="primary">
                                Join {orgDisplayName}
                            </Button>
                            <Button
                                disabled={loading}
                                className={styles.declineButton}
                                onClick={declineInvitation}
                                variant="secondary"
                                outline={true}
                            >
                                Decline
                            </Button>
                        </div>
                        {data.isVerifiedEmail === false && data.recipientEmail && (
                            <small className="mt-4 text-muted d-inline-block">
                                <AnchorLink to="/-/sign-out">Or sign out and create a new account</AnchorLink>
                                <br />
                                to join the {orgDisplayName} organization
                            </small>
                        )}
                    </Form>
                </ModalPage>
            )}
            {error && (
                <ModalPage className={classNames(styles.orgInvitationPage, className, 'p-4')}>
                    <Typography.H3>You've been invited to join an organization.</Typography.H3>
                    <Alert variant="danger" className="mt-3">
                        Error: {error}
                    </Alert>
                </ModalPage>
            )}
            {loading && <LoadingSpinner />}
        </>
    )
}
