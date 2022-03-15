import classNames from 'classnames'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { MarketingBlock } from '@sourcegraph/web/src/components/MarketingBlock'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Link, PageHeader } from '@sourcegraph/wildcard'

import { OrgAreaOrganizationFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgGetStartedInfo } from '../area/OrgArea'
import { OrgAreaHeaderContext } from '../area/OrgHeader'

import styles from './GettingStarted.module.scss'

export const showGetStartPage = (context: OrgAreaHeaderContext): boolean =>
    context.getStartedInfo.openBetaEnabled &&
    ((context.getStartedInfo.membersCount === 1 && context.getStartedInfo.invitesCount === 0) ||
        context.getStartedInfo.reposCount === 0 ||
        context.getStartedInfo.servicesCount === 0)

interface Props extends RouteComponentProps {
    authenticatedUser: AuthenticatedUser
    getStartedInfo: OrgGetStartedInfo
    org: OrgAreaOrganizationFields
}

const Step: React.FunctionComponent<{ completeCondition: boolean; textMuted: boolean; label: string; to?: string }> = ({
    completeCondition,
    label,
    textMuted,
    to,
}) => (
    <li className={styles.entryItem}>
        <div className={styles.iconContainer}>
            <CheckCircleIcon
                className={classNames('icon-inline', completeCondition ? 'text-success' : styles.iconMuted)}
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
        {!completeCondition && to && (
            <div className={styles.linkContainer}>
                <Link to={to}>
                    <ArrowRightIcon />
                </Link>
            </div>
        )}
    </li>
)

export const OpenBetaGetStartedPage: React.FunctionComponent<Props> = ({ authenticatedUser, getStartedInfo, org }) => {
    useEffect(() => {
        eventLogger.log('OpenBeta getting started')
    }, [])

    const codeHostsCompleted = getStartedInfo.servicesCount > 0
    const repoCompleted = getStartedInfo.reposCount > 0
    const membersCompleted = getStartedInfo.membersCount > 1 || getStartedInfo.invitesCount > 0
    const allCompleted = codeHostsCompleted && repoCompleted && membersCompleted
    return (
        <div className="org-members-page">
            <PageTitle title="Welcome to Sourcegraph!" />
            <div className="d-flex flex-column flex-0 justify-content-center align-items-center mb-3 flex-wrap">
                <PageHeader
                    path={[{ text: 'Welcome to Sourcegraph!' }]}
                    headingElement="h2"
                    className="mt-4 mb-4"
                    description={<span className="text-muted">Next, let’s get your organization up and running.</span>}
                />
                <MarketingBlock contentClassName={styles.boxContainer}>
                    <ul className={styles.entryItems}>
                        <Step
                            label="Connect with code hosts"
                            completeCondition={codeHostsCompleted}
                            textMuted={codeHostsCompleted}
                            to={`/organizations/${org.name}/settings/code-hosts`}
                        />
                        <Step
                            label="Choose repositories to sync with Sourcegraph"
                            completeCondition={repoCompleted}
                            textMuted={repoCompleted}
                            to={`/organizations/${org.name}/settings/repositories`}
                        />
                        <Step
                            label="Invite your teammates"
                            completeCondition={membersCompleted}
                            textMuted={membersCompleted}
                            to={`/organizations/${org.name}/settings/members`}
                        />
                        <Step
                            label="Search across Acmecorp’s code"
                            completeCondition={allCompleted}
                            to={allCompleted ? '#' : undefined}
                            textMuted={!allCompleted}
                        />
                    </ul>
                </MarketingBlock>
            </div>
        </div>
    )
}
