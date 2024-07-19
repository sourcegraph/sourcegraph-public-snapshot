import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { useNavigate, useParams } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { OrganizationInvitationResponseType } from '@sourcegraph/shared/src/graphql-operations'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Alert, AnchorLink, Button, Form, H2, H3, Link, LoadingSpinner } from '@sourcegraph/wildcard'

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
import { userURL } from '../../user'
import { OrgAvatar } from '../OrgAvatar'

import styles from './OrgInvitationPage.module.scss'

interface Props extends TelemetryV2Props {
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
    telemetryRecorder,
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
    const org = data?.organization
    const sender = data?.sender
    const willVerifyEmail = data?.recipientEmail && !data?.isVerifiedEmail

    useEffect(() => {
        if (org) {
            EVENT_LOGGER.logPageView('OrganizationInvitation', { organizationId: org.id, invitationId: data?.id })
            telemetryRecorder.recordEvent('org.invite', 'view')
        }
    }, [data?.id, telemetryRecorder, org])

    const [respondToInvitation, { loading: respondLoading, error: respondError }] = useMutation<
        RespondToOrgInvitationResult,
        RespondToOrgInvitationVariables
    >(RESPOND_TO_ORG_INVITATION, {
        onError: apolloError => {
            logger.error('Error when responding to invitation', apolloError)
        },
    })

    const acceptInvitation = useCallback(async () => {
        EVENT_LOGGER.log(
            'OrganizationInvitationAcceptClicked',
            {
                organizationId: org!.id,
                invitationId: data?.id,
                willVerifyEmail,
            },
            {
                organizationId: org!.id,
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
            EVENT_LOGGER.log(
                'OrganizationInvitationAcceptSucceeded',
                { organizationId: org!.id, invitationId: data?.id },
                { organizationId: org!.id, invitationId: data?.id }
            )
            telemetryRecorder.recordEvent('org.invite', 'accept')
        } catch {
            EVENT_LOGGER.log(
                'OrganizationInvitationAcceptFailed',
                { organizationId: org!.id, invitationId: data?.id },
                { organizationId: org!.id, invitationId: data?.id }
            )
            telemetryRecorder.recordEvent('org.invite', 'acceptFailed')
            return
        }

        navigate(orgURL(org!.name))
    }, [data?.id, navigate, org, respondToInvitation, willVerifyEmail, telemetryRecorder])

    const declineInvitation = useCallback(async () => {
        EVENT_LOGGER.log(
            'OrganizationInvitationDeclineClicked',
            {
                organizationId: org!.id,
                invitationId: data?.id,
                willVerifyEmail,
            },
            {
                organizationId: org!.id,
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
            EVENT_LOGGER.log(
                'OrganizationInvitationDeclineSucceeded',
                { organizationId: org!.id, invitationId: data?.id },
                { organizationId: org!.id, invitationId: data?.id }
            )
            telemetryRecorder.recordEvent('org.invite', 'decline')
        } catch {
            EVENT_LOGGER.log(
                'OrganizationInvitationDeclineFailed',
                { organizationId: org!.id, invitationId: data?.id },
                { organizationId: org!.id, invitationId: data?.id }
            )
            telemetryRecorder.recordEvent('org.invite', 'declineFailed')
        }

        navigate(userURL(authenticatedUser.username))
    }, [org, data?.id, willVerifyEmail, navigate, authenticatedUser.username, respondToInvitation, telemetryRecorder])

    const loading = inviteLoading || respondLoading
    const error = inviteError?.message || respondError?.message

    return (
        <>
            <PageTitle title={`Invitation to Organization ${org?.name || ''}`} />
            {org && sender && (
                <ModalPage
                    className={classNames(styles.orgInvitationPage, className)}
                    icon={<OrgAvatar org={org.name} className="mt-3 mb-4" size="lg" />}
                >
                    <Form className="text-center pr-4 pl-4 pb-4">
                        <H2>
                            You've been invited to join the <strong>{org.name}</strong>{' '}
                            {org.displayName ? `(${org.displayName})` : ''} organization
                        </H2>
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
                                <strong>{org.name}</strong> organization will add this as a verified email on your
                                account.
                            </div>
                        )}
                        <div className="mt-4">
                            <Button className="mr-sm-2" disabled={loading} onClick={acceptInvitation} variant="primary">
                                Join {org.name}
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
                                to join the {org.name} organization
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
