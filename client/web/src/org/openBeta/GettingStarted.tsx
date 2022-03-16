import React, { useEffect } from 'react'

import { gql, useQuery } from '@apollo/client'
import classNames from 'classnames'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { MarketingBlock } from '@sourcegraph/web/src/components/MarketingBlock'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import {
    GetStartedPageDataResult,
    GetStartedPageDataVariables,
    OrgAreaOrganizationFields,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgSummary } from '../area/OrgHeader'
import { Member } from '../members/OrgMembersListPage'

import styles from './GettingStarted.module.scss'

const GET_STARTED_STATUS = 'sgGetStarted'

const GET_STARTED_INFO_QUERY = gql`
    query GetStartedPageData($organization: ID!) {
        membersSummary: orgMembersSummary(organization: $organization) {
            membersCount
            invitesCount
        }
        repoCount: node(id: $organization) {
            ... on Org {
                total: repositories(cloned: true, notCloned: true) {
                    totalCount(precise: true)
                }
            }
        }
        extServices: externalServices(namespace: $organization) {
            totalCount
        }
        membersList: node(id: $organization) {
            ... on Org {
                members {
                    nodes {
                        id
                        username
                        displayName
                        avatarURL
                    }
                }
            }
        }
    }
`

export const showGetStartPage = (info: OrgSummary | undefined, orgName: string, openBetaEnabled: boolean): boolean => {
    if (!openBetaEnabled) {
        return false
    }

    if (!info) {
        return true
    }

    const firstStepsPending =
        (info.membersSummary.membersCount === 1 && info.membersSummary.invitesCount === 0) ||
        info.repoCount.total.totalCount === 0 ||
        info.extServices.totalCount === 0
    let searchStatusPending = true
    try {
        const status = localStorage.getItem(`${GET_STARTED_STATUS}${orgName}`)
        searchStatusPending = !!status && status !== 'complete'
    } catch {
        eventLogger.log('getStartedStatusError')
    }
    return firstStepsPending || searchStatusPending
}

interface Props extends RouteComponentProps, FeatureFlagProps {
    authenticatedUser: AuthenticatedUser
    org: OrgAreaOrganizationFields
}

const Step: React.FunctionComponent<{ complete: boolean; textMuted: boolean; label: string; to?: string }> = ({
    complete,
    label,
    textMuted,
    to,
}) => (
    <li className={styles.entryItem}>
        <div className={styles.iconContainer}>
            <CheckCircleIcon
                className={classNames('icon-inline', complete ? 'text-success' : styles.iconMuted)}
                size="1rem"
            />
        </div>
        <h3
            className={classNames({
                [`${styles.stepText}`]: true,
                'text-muted': textMuted,
            })}
        >
            {label}
        </h3>
        {to && (
            <div className={styles.linkContainer}>
                <Link to={to}>
                    <ArrowRightIcon />
                </Link>
            </div>
        )}
    </li>
)

const InviteLink: React.FunctionComponent<{ orgName: string; membersCount: number }> = ({ membersCount, orgName }) => {
    const preText = membersCount === 1 ? 'It’s just you so far! ' : null
    const linkText = membersCount === 1 ? 'Invite your teammates' : 'Invite the rest of your teammates'
    return (
        <small>
            {preText}
            <Link to={`/organizations/${orgName}/settings/members/pending-invites?openInviteModal=1&openBetaBanner=1`}>
                {linkText}
            </Link>
        </small>
    )
}

export const OpenBetaGetStartedPage: React.FunctionComponent<Props> = ({
    authenticatedUser,
    org,
    featureFlags,
    history,
}) => {
    const openBetaEnabled = !!featureFlags.get('open-beta-enabled')
    const { data, loading, error } = useQuery<GetStartedPageDataResult, GetStartedPageDataVariables>(
        GET_STARTED_INFO_QUERY,
        {
            variables: { organization: org.id },
        }
    )

    const queryResult = data ? (data as OrgSummary & { membersList: { members: { nodes: Member[] } } }) : undefined

    const searchStepStorageKey = `${GET_STARTED_STATUS}${org.name}`
    const codeHostsCompleted = !!queryResult && queryResult.extServices.totalCount > 0
    const repoCompleted = !!queryResult && queryResult.repoCount.total.totalCount > 0
    const membersCompleted =
        !!queryResult && (queryResult.membersSummary.membersCount > 1 || queryResult.membersSummary.invitesCount > 0)
    const allCompleted = codeHostsCompleted && repoCompleted && membersCompleted
    const searchCompleted = codeHostsCompleted && repoCompleted
    const membersResult = queryResult ? queryResult.membersList.members.nodes : []
    const otherMembers =
        membersResult.length > 1 ? membersResult.filter(user => user.username !== authenticatedUser.username) : []
    const shouldRedirect =
        !openBetaEnabled || (queryResult && !showGetStartPage(queryResult, org.name, openBetaEnabled))

    useEffect(() => {
        eventLogger.log('OpenBeta getting started')
    }, [])

    useEffect(() => {
        if (queryResult && !loading) {
            const currentSearchStatus = localStorage.getItem(searchStepStorageKey)
            const newStatus = !allCompleted
                ? 'incomplete'
                : currentSearchStatus === 'incomplete'
                ? 'active'
                : 'complete'
            localStorage.setItem(searchStepStorageKey, newStatus)
        }
    }, [allCompleted, searchStepStorageKey, queryResult, loading])

    useEffect(() => {
        if (shouldRedirect) {
            history.push(`/organizations/${org.name}/settings/members`)
        }
    }, [shouldRedirect, org.name, history])

    if (shouldRedirect) {
        return null
    }

    return (
        <>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            {queryResult && !loading && (
                <div className="org-getstarted-page">
                    <PageTitle title="Welcome to Sourcegraph!" />
                    <div className="d-flex flex-column flex-0 justify-content-center align-items-center mb-3 flex-wrap">
                        <PageHeader
                            path={[{ text: 'Welcome to Sourcegraph!' }]}
                            headingElement="h2"
                            className="mt-4 mb-4"
                            description={
                                <span className="text-muted">Next, let’s get your organization up and running.</span>
                            }
                        />

                        <MarketingBlock contentClassName={styles.boxContainer}>
                            <ul className={styles.entryItems}>
                                <Step
                                    label="Connect with code hosts"
                                    complete={codeHostsCompleted}
                                    textMuted={codeHostsCompleted}
                                    to={
                                        codeHostsCompleted
                                            ? undefined
                                            : `/organizations/${org.name}/settings/code-hosts`
                                    }
                                />
                                <Step
                                    label="Choose repositories to sync with Sourcegraph"
                                    complete={repoCompleted}
                                    textMuted={repoCompleted}
                                    to={repoCompleted ? undefined : `/organizations/${org.name}/settings/repositories`}
                                />
                                <Step
                                    label="Invite your teammates"
                                    complete={membersCompleted}
                                    textMuted={membersCompleted}
                                    to={
                                        membersCompleted
                                            ? undefined
                                            : `/organizations/${org.name}/settings/members/pending-invites?openInviteModal=1&openBetaBanner=1`
                                    }
                                />
                                <Step
                                    label={`Search across ${org.displayName || org.name}’s code`}
                                    complete={false}
                                    to={
                                        searchCompleted
                                            ? `/search?q=context:%40${org.name}&patternType=literal`
                                            : undefined
                                    }
                                    textMuted={!searchCompleted}
                                />
                            </ul>
                        </MarketingBlock>

                        <div className="mt-4">
                            <div className="d-flex  flex-0 justify-content-center align-items-center mb-3 flex-wrap">
                                <div className={styles.membersList}>
                                    <div className={styles.avatarContainer}>
                                        <UserAvatar
                                            size={36}
                                            className={styles.avatar}
                                            user={authenticatedUser}
                                            data-tooltip={authenticatedUser.displayName || authenticatedUser.username}
                                        />
                                    </div>
                                    {otherMembers.length > 0 && (
                                        <div className={styles.avatarContainer}>
                                            <div className={classNames(styles.avatarEllipse)} />
                                            <div className={classNames(styles.avatarContainer, styles.secondAvatar)}>
                                                <UserAvatar
                                                    size={36}
                                                    className={styles.avatar}
                                                    user={otherMembers[0]}
                                                    data-tooltip={
                                                        otherMembers[0].displayName || otherMembers[0].username
                                                    }
                                                />
                                            </div>
                                        </div>
                                    )}
                                    {otherMembers.length > 1 && (
                                        <div className={styles.avatarContainer}>
                                            <div
                                                className={classNames(styles.avatarEllipse, styles.avatarEllipseSecond)}
                                            />
                                            <div className={classNames(styles.totalCount, 'text-muted')}>{`+${
                                                queryResult.membersSummary.membersCount - 2
                                            }`}</div>
                                        </div>
                                    )}
                                </div>
                                <InviteLink orgName={org.name} membersCount={queryResult.membersSummary.membersCount} />
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </>
    )
}
