import React, { useState, useEffect, useMemo } from 'react'

import { mdiPlusThick, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Icon, PageHeader, Button, Link, Text, H3, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { fetchThroughSSCProxy } from '../util'

import { InviteUsers } from './InviteUsers'
import { type TeamInvite, TeamMemberList, type TeamMember } from './TeamMembers'
import { WhiteIcon } from './WhiteIcon'

import styles from './CodyManageTeamPage.module.scss'

interface CodyManageTeamPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

// TODO: Remove this mock data
const mockTeamMembers: TeamMember[] = [
    {
        accountId: '1',
        displayName: 'daniel.marques.pt',
        email: 'daniel.marques@sourcegraph.com',
        avatarUrl: null,
        role: 'admin',
    },
]

const mockInvites: TeamInvite[] = [
    {
        id: '1',
        email: 'rob.rhyne@sourcegraph.com',
        role: 'member',
        status: 'sent',
        error: null,
        sentAt: '2021-09-01T00:00:00Z',
        acceptedAt: null,
    },
    {
        id: '2',
        email: 'kevin.chen@sourcegraph.com',
        role: 'admin',
        status: 'sent',
        error: null,
        sentAt: '2021-09-01T00:00:00Z',
        acceptedAt: null,
    },
    {
        id: '3',
        email: 'test@test.com',
        role: 'member',
        status: 'accepted',
        error: null,
        sentAt: '2021-09-01T00:00:00Z',
        acceptedAt: '2021-09-01T00:00:00Z',
    },
]

export const CodyManageTeamPage: React.FunctionComponent<CodyManageTeamPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.team.management', 'view')
    }, [telemetryRecorder])

    const navigate = useNavigate()

    // TODO: Make this dynamic
    const isAdmin = true

    // Process query params
    const parameters = useSearchParameters()
    const newSeatsPurchasedParam = parameters.get('newSeatsPurchased')
    const newSeatsPurchased: number | null = newSeatsPurchasedParam ? parseInt(newSeatsPurchasedParam, 10) : null

    // Load data
    const [subscriptionData, setSubscriptionData] = useState<{
        subscriptionSeatCount: number | null
        isProUser: boolean | null
    } | null>(null)
    const subscriptionSeatCount = subscriptionData?.subscriptionSeatCount
    const isProUser = subscriptionData?.isProUser
    const [subscriptionDataError, setSubscriptionDataError] = useState<null | Error>(null)
    const [teamId, setTeamId] = useState<string | null>(null)
    // const [teamDataError, setTeamDataError] = useState<null | Error>(null)
    const [teamMembers, setTeamMembers] = useState<TeamMember[] | null>(null)
    const [membersDataError, setMembersDataError] = useState<null | Error>(null)
    const [teamInvites, setTeamInvites] = useState<TeamInvite[] | null>(null)
    const [invitesDataError, setInvitesDataError] = useState<null | Error>(null)
    useEffect(() => {
        // async function loadTeamData(): Promise<void> {
        //     try {
        //         const response = await fetchThroughSSCProxy('/team/current', 'GET')
        //         const responseJson = await response.json()
        //         setTeamId((responseJson as { id: string }).id)
        //     } catch (error) {
        //         setTeamDataError(error)
        //     }
        // }
        async function loadSubscriptionData(): Promise<void> {
            try {
                const response = await fetchThroughSSCProxy('/team/current/subscription', 'GET')
                const responseJson = (await response.json()) as {
                    subscriptionStatus: 'active' | 'past_due' | 'unpaid' | 'canceled' | 'trialing' | 'other'
                    maxSeats: number
                } | null
                setSubscriptionData({
                    subscriptionSeatCount: responseJson?.maxSeats ?? null,
                    isProUser: responseJson && responseJson.subscriptionStatus !== 'canceled',
                })
            } catch (error) {
                setSubscriptionDataError(error)
            }
        }
        async function loadMemberData(): Promise<void> {
            try {
                const response = await fetchThroughSSCProxy('/team/current/members', 'GET')
                const responseJson = await response.json()
                setTeamMembers((responseJson as { members: TeamMember[] }).members.concat(mockTeamMembers))
            } catch (error) {
                setMembersDataError(error)
            }
        }
        async function loadInviteData(): Promise<void> {
            try {
                const response = await fetchThroughSSCProxy('/team/current/invites', 'GET')
                const responseJson = await response.json()
                setTeamInvites((responseJson as { invites: TeamInvite[] }).invites.concat(mockInvites))
            } catch (error) {
                setInvitesDataError(error)
            }
        }

        // TODO: Make this live once we have the endpoint
        // void loadTeamData()
        setTeamId('1')

        void loadSubscriptionData()
        void loadMemberData()
        void loadInviteData()
    }, [authenticatedUser])

    useEffect(() => {
        if (isProUser === false) {
            navigate('/cody/subscription')
        }
    }, [isProUser, navigate])

    const remainingInviteCount = useMemo(() => {
        const memberCount = teamMembers?.length ?? 0
        const invitesUsed = (teamInvites ?? []).filter(invite => invite.status === 'sent').length
        return Math.max((subscriptionSeatCount ?? 0) - (memberCount + invitesUsed), 0)
    }, [subscriptionSeatCount, teamMembers, teamInvites])

    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="Manage Cody team" />
                <PageHeader
                    className="mb-4 mt-4"
                    actions={
                        isAdmin && (
                            <div className="d-flex">
                                <Link
                                    to="/cody/manage"
                                    className="d-inline-flex align-items-center mr-3"
                                    onClick={() =>
                                        telemetryRecorder.recordEvent('cody.team.manage.subscription', 'click', {
                                            metadata: { tier: isProUser ? 1 : 0 },
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

                {subscriptionDataError || membersDataError || invitesDataError ? (
                    <div className={classNames('mb-4', styles.alert, styles.errorAlert)}>
                        <H3>Failed to load team data.</H3>
                        {subscriptionDataError?.message && (
                            <Text size="small" className="text-muted mb-0">
                                {subscriptionDataError?.message}
                            </Text>
                        )}
                        {membersDataError?.message && (
                            <Text size="small" className="text-muted mb-0">
                                {membersDataError?.message}
                            </Text>
                        )}
                        {invitesDataError?.message && (
                            <Text size="small" className="text-muted mb-0">
                                {invitesDataError?.message}
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
                        teamId={teamId}
                        remainingInviteCount={remainingInviteCount}
                        telemetryRecorder={telemetryRecorder}
                    />
                )}
                <TeamMemberList
                    teamId={teamId}
                    teamMembers={teamMembers || []}
                    invites={teamInvites || []}
                    isAdmin={isAdmin}
                    telemetryRecorder={telemetryRecorder}
                />
            </Page>
        </>
    )
}
