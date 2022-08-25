import React, { ReactNode, useEffect } from 'react'

import { gql, useQuery } from '@apollo/client'
import { mdiCheckCircle, mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link, LoadingSpinner, PageHeader, Badge, H3, Icon, Tooltip } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import {
    GetStartedPageDataResult,
    GetStartedPageDataVariables,
    OrgAreaOrganizationFields,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { OrgSummary } from '../area/OrgHeader'
import { useEventBus } from '../emitter'
import { Member } from '../members/OrgMembersListPage'

import styles from './GettingStarted.module.scss'

const GET_STARTED_INFO_QUERY = gql`
    query GetStartedPageData($organization: ID!) {
        membersSummary: orgMembersSummary(organization: $organization) {
            membersCount
            invitesCount
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

export const calculateLeftGetStartedSteps = (info: OrgSummary | undefined, orgName: string): ReactNode => {
    if (!info) {
        return 4
    }

    let leftSteps = 0
    if (info.membersSummary.invitesCount === 0 && info.membersSummary.membersCount < 2) {
        leftSteps += 1
    }
    if (info.extServices.totalCount === 0) {
        setSearchStep(orgName, 'incomplete')
        leftSteps += 1
    }

    const searchPending = !isSearchStepComplete(orgName)
    if (searchPending) {
        leftSteps += 1
    }

    return (
        <span>
            Get started{' '}
            <Badge pill={true} className={styles.badge} variant="secondary">
                {leftSteps}
            </Badge>
        </span>
    )
}

export const showGetStartPage = (
    info: OrgSummary | undefined,
    orgName: string,
    openBetaEnabled: boolean,
    isDotCom: boolean
): boolean => {
    if (!openBetaEnabled || !isDotCom) {
        return false
    }

    if (!info) {
        return false
    }

    const firstStepsPending =
        (info.membersSummary.membersCount === 1 && info.membersSummary.invitesCount === 0) ||
        info.extServices.totalCount === 0
    let searchStatusPending = true
    try {
        searchStatusPending = !isSearchStepComplete(orgName)
    } catch {
        eventLogger.log('getStartedStatusError')
    }
    return firstStepsPending || searchStatusPending
}

interface Props extends RouteComponentProps {
    authenticatedUser: AuthenticatedUser
    org: OrgAreaOrganizationFields
    isSourcegraphDotCom: boolean
}

const LinkableContainer: React.FunctionComponent<React.PropsWithChildren<{ to?: string; onClick?: () => void }>> = ({
    to,
    onClick,
    children,
}) => {
    if (to) {
        return (
            <Link className={styles.entryItemLink} to={to} onClick={onClick}>
                {children}
            </Link>
        )
    }

    return <>{children}</>
}

const Step: React.FunctionComponent<
    React.PropsWithChildren<{
        complete: boolean
        textMuted: boolean
        label: string
        to?: string
        onClick?: () => void
    }>
> = ({ complete, label, textMuted, to, onClick }) => (
    <li className={styles.entryItem}>
        <LinkableContainer to={to} onClick={onClick}>
            <div className={styles.iconContainer}>
                {complete && (
                    <Icon
                        className="text-success"
                        svgPath={mdiCheckCircle}
                        inline={false}
                        aria-label="Success"
                        height={14}
                        width={14}
                    />
                )}
                {!complete && <div className={styles.emptyCircle} />}
            </div>
            <H3
                className={classNames({
                    [`${styles.stepText}`]: true,
                    'text-muted': textMuted,
                })}
            >
                {label}
            </H3>
            {to && (
                <div className={styles.linkContainer}>
                    <Icon svgPath={mdiArrowRight} inline={false} aria-hidden={true} />
                </div>
            )}
        </LinkableContainer>
    </li>
)

const InviteLink: React.FunctionComponent<
    React.PropsWithChildren<{ orgName: string; orgId: string; membersCount: number }>
> = ({ membersCount, orgId, orgName }) => {
    const preText = membersCount === 1 ? 'It’s just you so far! ' : null
    const linkText = membersCount === 1 ? 'Invite your teammates' : 'Invite the rest of your teammates'
    return (
        <small>
            {preText}
            <Link
                to={`/organizations/${orgName}/settings/members/pending-invites?openInviteModal=1&openBetaBanner=1`}
                onClick={() =>
                    eventLogger.log(
                        'OrganizationGetStartedInviteTeammatesCTAClicked',
                        { organizationId: orgId },
                        { organizationId: orgId }
                    )
                }
            >
                {linkText}
            </Link>
        </small>
    )
}

const SEARCH_STATUS_RADIX = 'sgGetStartedSearchStep'
const getSearchStepStatus = (orgName: string): 'complete' | 'incomplete' | null => {
    try {
        return localStorage.getItem(`${SEARCH_STATUS_RADIX}${orgName}`) as 'complete' | 'incomplete' | null
    } catch (error) {
        eventLogger.log('GetStartedLocalStorageError', error)
        return null
    }
}
const isSearchStepComplete = (orgName: string): boolean => {
    const stepStatus = getSearchStepStatus(orgName)
    return !stepStatus || stepStatus === 'complete'
}
const setSearchStep = (orgName: string, status: 'complete' | 'incomplete'): void => {
    try {
        localStorage.setItem(`${SEARCH_STATUS_RADIX}${orgName}`, status)
    } catch (error) {
        eventLogger.log('GetStartedLocalStorageError', error)
    }
}

export const OpenBetaGetStartedPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    org,
    history,
    isSourcegraphDotCom,
}) => {
    const emitter = useEventBus()

    const [isOpenBetaEnabled] = useFeatureFlag('open-beta-enabled')
    const { data, loading, error } = useQuery<GetStartedPageDataResult, GetStartedPageDataVariables>(
        GET_STARTED_INFO_QUERY,
        {
            variables: { organization: org.id },
        }
    )

    const queryResult = data ? (data as OrgSummary & { membersList: { members: { nodes: Member[] } } }) : undefined

    const codeHostsCompleted = !!queryResult && queryResult.extServices.totalCount > 0
    // TODO: This is set to false after removal of org-owned repositories. Once we remove more of this, we can remove the whole file
    const repoCompleted = false
    const membersCompleted =
        !!queryResult && (queryResult.membersSummary.membersCount > 1 || queryResult.membersSummary.invitesCount > 0)
    const allowSearch = codeHostsCompleted && repoCompleted
    const membersResult = queryResult ? queryResult.membersList.members.nodes : []
    const otherMembers =
        membersResult.length > 1 ? membersResult.filter(user => user.username !== authenticatedUser.username) : []
    const shouldRedirect =
        !isOpenBetaEnabled ||
        (queryResult && !showGetStartPage(queryResult, org.name, isOpenBetaEnabled, isSourcegraphDotCom))

    useEffect(() => {
        eventLogger.logPageView('OrganizationGetStarted', { organizationId: org.id })
    }, [org.id])

    useEffect(() => {
        if (queryResult && !loading && !allowSearch) {
            setSearchStep(org.name, 'incomplete')
            emitter.emit('refreshOrgHeader', 'changedsearchPrerequisites')
        }
    }, [allowSearch, org.name, queryResult, loading, emitter])

    useEffect(() => {
        if (shouldRedirect) {
            history.push(`/organizations/${org.name}/settings/members`)
        }
    }, [shouldRedirect, org.name, history])

    const completeSearchStep = (): void => {
        setSearchStep(org.name, 'complete')
        emitter.emit('refreshOrgHeader', 'searchdone')
    }

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
                            className={classNames('mt-4 mb-4 justify-content-center', styles.headingTitle)}
                            description={
                                <span className="text-muted">Next, let’s get your organization up and running.</span>
                            }
                        />

                        <MarketingBlock
                            contentClassName={styles.boxContainer}
                            wrapperClassName={styles.boxContainerWrapper}
                        >
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
                                    onClick={() =>
                                        eventLogger.log(
                                            'OrganizationGetStartedCodeHostsClicked',
                                            { organizationId: org.id },
                                            { organizationId: org.id }
                                        )
                                    }
                                />
                                <Step
                                    label="Choose repositories to sync with Sourcegraph"
                                    complete={repoCompleted}
                                    textMuted={repoCompleted}
                                    to={repoCompleted ? undefined : `/organizations/${org.name}/settings/repositories`}
                                    onClick={() =>
                                        eventLogger.log(
                                            'OrganizationGetStartedChooseRepositoriesClicked',
                                            { organizationId: org.id },
                                            { organizationId: org.id }
                                        )
                                    }
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
                                    onClick={() =>
                                        eventLogger.log(
                                            'OrganizationGetStartedInviteTeammatesClicked',
                                            { organizationId: org.id },
                                            { organizationId: org.id }
                                        )
                                    }
                                />
                                <Step
                                    label={`Search across ${org.displayName || org.name}’s code`}
                                    complete={getSearchStepStatus(org.name) === 'complete'}
                                    onClick={() => {
                                        eventLogger.log(
                                            'OrganizationGetStartedSearchClicked',
                                            { organizationId: org.id },
                                            { organizationId: org.id }
                                        )
                                        completeSearchStep()
                                    }}
                                    to={
                                        allowSearch ? `/search?q=context:%40${org.name}&patternType=literal` : undefined
                                    }
                                    textMuted={!allowSearch}
                                />
                            </ul>
                        </MarketingBlock>

                        <div className="mt-4">
                            <div className="d-flex  flex-0 justify-content-center align-items-center mb-3 flex-wrap">
                                <div className={styles.membersList}>
                                    <div className={styles.avatarContainer}>
                                        <Tooltip content={authenticatedUser.displayName || authenticatedUser.username}>
                                            <UserAvatar size={36} className={styles.avatar} user={authenticatedUser} />
                                        </Tooltip>
                                    </div>
                                    {otherMembers.length > 0 && (
                                        <div className={styles.avatarContainer}>
                                            <div className={classNames(styles.avatarEllipse)} />
                                            <div className={classNames(styles.avatarContainer, styles.secondAvatar)}>
                                                <Tooltip
                                                    content={otherMembers[0].displayName || otherMembers[0].username}
                                                >
                                                    <UserAvatar
                                                        size={36}
                                                        className={styles.avatar}
                                                        user={otherMembers[0]}
                                                    />
                                                </Tooltip>
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
                                <InviteLink
                                    orgName={org.name}
                                    orgId={org.id}
                                    membersCount={queryResult.membersSummary.membersCount}
                                />
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </>
    )
}
