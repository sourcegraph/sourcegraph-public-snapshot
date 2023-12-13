import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { useNavigate, useParams } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { OrganizationInvitationResponseType } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, AnchorLink, Button, LoadingSpinner, Link, H2, H3, Form } from '@sourcegraph/wildcard'

import { orgURL } from '..'
import type { AuthenticatedUser } from '../../auth'
import { ModalPage } from '../../components/ModalPage'
import { PageTitle } from '../../components/PageTitle'
import type {
    InvitationByTokenResult,
    InvitationByTokenVariables,
    RespondToOrgInvitationResult,
    RespondToOrgInvitationVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { OrgAvatar } from '../OrgAvatar'

import styles from './OrgInvitationPage.module.scss'

interface Props {
    authenticatedUser: AuthenticatedUser
    className?: string
}

export const RESPOND_TO_ORG_INVITATION = gql`
    mutation RespondToOrgInvitation($id: ID!, $response: OrganizationInvitationResponseType!) {
        respondToOrganizationInvitation(organizationInvitation: $id, responseType: $response) {
            alwaysNil
        }
    }
`

export const INVITATION_BY_TOKEN = gql`
    query InvitationByToken($token: String!) {
        invitationByToken(token: $token) {
            ...OrganizationInvitationFields
        }
    }

    fragment OrganizationInvitationFields on OrganizationInvitation {
        createdAt
        id
        isVerifiedEmail
        organization {
            id
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
`

/**
 * Displays the organization invitation for the user, based on the token in the invite URL.
 */
export const OrgInvitationPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    className,
}) => {
    const { token } = useParams<{ token: string }>()
    const navigate = useNavigate()

    const {
        data: inviteData,
        loading: inviteLoading,
        error: inviteError,
    } = useQuery<InvitationByTokenResult, InvitationByTokenVariables>(INVITATION_BY_TOKEN, {
        skip: !authenticatedUser || !token,
        variables: {
            token: token!,
        },
    })

    const data = inviteData?.invitationByToken
    const orgName = data?.organization.name
    const orgId = data?.organization.id
    const sender = data?.sender
    const orgDisplayName = data?.organization.displayName || orgName
    const willVerifyEmail = data?.recipientEmail && !data?.isVerifiedEmail

    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('organizationInvitation', 'viewed', {
            privateMetadata: { organizationId: orgId, invitationId: data?.id },
        })
        eventLogger.logPageView('OrganizationInvitation', { organizationId: orgId, invitationId: data?.id })
    }, [orgId, data?.id])

    const [respondToInvitation, { loading: respondLoading, error: respondError }] = useMutation<
        RespondToOrgInvitationResult,
        RespondToOrgInvitationVariables
    >(RESPOND_TO_ORG_INVITATION, {
        onError: apolloError => {
            logger.error('Error when responding to invitation', apolloError)
        },
    })

    const acceptInvitation = useCallback(async () => {
        window.context.telemetryRecorder?.recordEvent('organizationInvitation', 'accepted', {
            privateMetadata: { organizationId: orgId, invitationId: data?.id, willVerifyEmail },
        })
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
            window.context.telemetryRecorder?.recordEvent('organizationInvitation', 'accepted', {
                privateMetadata: { organizationId: orgId, invitationId: data?.id },
            })
            eventLogger.log(
                'OrganizationInvitationAcceptSucceeded',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        } catch {
            window.context.telemetryRecorder?.recordEvent('organizationInvitation', 'rejected', {
                privateMetadata: { organizationId: orgId, invitationId: data?.id },
            })
            eventLogger.log(
                'OrganizationInvitationAcceptFailed',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
            return
        }

        if (orgName) {
            navigate(orgURL(orgName))
        }
    }, [data?.id, navigate, orgId, orgName, respondToInvitation, willVerifyEmail])

    const declineInvitation = useCallback(async () => {
        window.context.telemetryRecorder?.recordEvent('organizationInvitationDecline', 'clicked', {
            privateMetadata: { organizationId: orgId, invitationId: data?.id, willVerifyEmail },
        })
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
            window.context.telemetryRecorder?.recordEvent('organizationInvitationDecline', 'succeeded', {
                privateMetadata: { organizationId: orgId, invitationId: data?.id },
            })
            eventLogger.log(
                'OrganizationInvitationDeclineSucceeded',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        } catch {
            window.context.telemetryRecorder?.recordEvent('organizationInvitationDecline', 'failed', {
                privateMetadata: { organizationId: orgId, invitationId: data?.id },
            })
            eventLogger.log(
                'OrganizationInvitationDeclineFailed',
                { organizationId: orgId, invitationId: data?.id },
                { organizationId: orgId, invitationId: data?.id }
            )
        }

        navigate(userURL(authenticatedUser.username))
    }, [authenticatedUser.username, data?.id, navigate, orgId, respondToInvitation, willVerifyEmail])

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
                        <H2>You've been invited to join the {orgDisplayName} organization</H2>
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
                    <H3>You've been invited to join an organization.</H3>
                    <Alert variant="danger" className="mt-3">
                        Error: {error}
                    </Alert>
                </ModalPage>
            )}
            {loading && <LoadingSpinner />}
        </>
    )
}
