import React, { useEffect, useMemo } from 'react'

import { mdiPlusThick, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Icon, PageHeader, Button, Link, Text, H3, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useCodySubscriptionSummaryData } from '../subscription/subscriptionSummary'
import { useSSCQuery } from '../util'

import { InviteUsers } from './InviteUsers'
import { TeamMemberList, type TeamMember, type TeamInvite } from './TeamMemberList'
import { WhiteIcon } from './WhiteIcon'

import styles from './CodyManageTeamPage.module.scss'

interface CodyManageTeamPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

type CodySubscriptionStatus = 'active' | 'past_due' | 'unpaid' | 'canceled' | 'trialing' | 'other'

interface CodySubscription {
    subscriptionStatus: CodySubscriptionStatus
    maxSeats: number
}

const AuthenticatedCodyManageTeamPage: React.FunctionComponent<CodyManageTeamPageProps> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.team.management', 'view')
    }, [telemetryRecorder])

    const navigate = useNavigate()

    // Process query params
    const parameters = useSearchParameters()
    const newSeatsPurchasedParam = parameters.get('newSeatsPurchased')
    const newSeatsPurchased: number | null = newSeatsPurchasedParam ? parseInt(newSeatsPurchasedParam, 10) : null

    // Load data
    const [codySubscription, codySubscriptionError] = useSSCQuery<CodySubscription>('/team/current/subscription')
    const isPro = codySubscription?.subscriptionStatus !== 'canceled'
    const [codySubscriptionSummary, codySubscriptionSummaryError] = useCodySubscriptionSummaryData()
    const isAdmin = codySubscriptionSummary?.userRole === 'admin'
    const [memberResponse, membersDataError] = useSSCQuery<{ members: TeamMember[] }>('/team/current/members')
    const teamMembers = memberResponse?.members
    const [invitesResponse, invitesDataError] = useSSCQuery<{ invites: TeamInvite[] }>('/team/current/invites')
    const teamInvites = invitesResponse?.invites
    const errorMessage =
        codySubscriptionError?.message ||
        codySubscriptionSummaryError?.message ||
        membersDataError?.message ||
        invitesDataError?.message

    useEffect(() => {
        if (!isPro) {
            navigate('/cody/subscription')
        }
    }, [isPro, navigate])

    const remainingInviteCount = useMemo(() => {
        const memberCount = teamMembers?.length ?? 0
        const invitesUsed = (teamInvites ?? []).filter(invite => invite.status === 'sent').length
        return Math.max((codySubscription?.maxSeats ?? 0) - (memberCount + invitesUsed), 0)
    }, [codySubscription?.maxSeats, teamMembers, teamInvites])

    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="Manage Cody team" />
                <PageHeader
                    className="mb-4 mt-4"
                    actions={
                        codySubscriptionSummary?.userRole === 'admin' && (
                            <div className="d-flex">
                                <Link
                                    to="/cody/manage"
                                    className="d-inline-flex align-items-center mr-3"
                                    onClick={() =>
                                        telemetryRecorder.recordEvent('cody.team.manage.subscription', 'click', {
                                            metadata: {
                                                tier: codySubscription?.subscriptionStatus !== 'canceled' ? 1 : 0,
                                            },
                                        })
                                    }
                                >
                                    Manage subscription
                                    <Icon
                                        svgPath={mdiOpenInNew}
                                        inline={false}
                                        aria-hidden={true}
                                        height="1rem"
                                        width="1rem"
                                        className="ml-2"
                                    />
                                </Link>
                                <Button
                                    as={Link}
                                    to="/cody/manage/subscription/new"
                                    variant="primary"
                                    className="text-nowrap"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiPlusThick} /> Add seats
                                </Button>
                            </div>
                        )
                    }
                >
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <div className="d-inline-flex align-items-center">
                            <WhiteIcon name="mdi-account-multiple-plus-gradient" />
                        </div>
                    </PageHeader.Heading>
                </PageHeader>

                {codySubscriptionError || codySubscriptionSummaryError || membersDataError || invitesDataError ? (
                    <div className={classNames('mb-4', styles.alert, styles.errorAlert)}>
                        <H3>We couldn't load team data this time. Please try a bit later.</H3>
                        {errorMessage ?? (
                            <Text size="small" className="text-muted mb-0">
                                {errorMessage}
                            </Text>
                        )}
                    </div>
                ) : null}

                {newSeatsPurchased && (
                    <div className={classNames('mb-4', styles.alert, styles.purpleSuccessAlert)}>
                        <H3>{newSeatsPurchased} Cody teams seats purchased!</H3>
                        <Text size="small" className="mb-0">
                            Invited users will receive unlimited autocompletions and unlimited chat messages.
                        </Text>
                    </div>
                )}

                {isAdmin && !!remainingInviteCount && (
                    <InviteUsers
                        teamId={codySubscriptionSummary?.teamId}
                        remainingInviteCount={remainingInviteCount}
                        telemetryRecorder={telemetryRecorder}
                    />
                )}
                <TeamMemberList
                    teamId={codySubscriptionSummary?.teamId ?? null}
                    teamMembers={teamMembers || []}
                    invites={teamInvites || []}
                    isAdmin={isAdmin}
                    telemetryRecorder={telemetryRecorder}
                />
            </Page>
        </>
    )
}

export const CodyManageTeamPage = withAuthenticatedUser(AuthenticatedCodyManageTeamPage)
