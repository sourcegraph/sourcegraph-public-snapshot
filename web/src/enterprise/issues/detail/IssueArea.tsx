import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { threadlikeSidebarSections } from '../../threadlike/sidebar/ThreadlikeSidebar'
import { IssueDeleteButton } from '../common/IssueDeleteButton'
import { RepositoryIssuesAreaContext } from '../repository/RepositoryIssuesArea'
import { IssueOverview } from './IssueOverview'
import { useIssueByNumberInRepository } from './useIssueByNumberInRepository'
import { IssueActivity } from './activity/IssueActivity'

export interface IssueAreaContext
    extends Pick<RepositoryIssuesAreaContext, Exclude<keyof RepositoryIssuesAreaContext, 'repository'>> {
    /** The issue, queried from the GraphQL API. */
    issue: GQL.IIssue

    /** Called to refresh the issue. */
    onIssueUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<IssueAreaContext, Exclude<keyof IssueAreaContext, 'issue' | 'onIssueUpdate'>>,
        RouteComponentProps<{}> {
    /**
     * The issue ID in its repository (i.e., the `Issue.number` GraphQL API field).
     */
    issueNumber: GQL.IIssue['number']

    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container mt-4'

/**
 * The area for a single issue.
 */
export const IssueArea: React.FunctionComponent<Props> = ({
    header,
    issueNumber,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [issue, onIssueUpdate] = useIssueByNumberInRepository(props.repo.id, issueNumber)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                issue !== LOADING && issue !== null && !isErrorLike(issue)
                    ? { text: `#${issue.number}`, to: issue.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [issue, setBreadcrumbItem])

    const onIssueDelete = useCallback(() => {
        if (issue !== LOADING && issue !== null && !isErrorLike(issue)) {
            props.history.push(`${issue.repository.url}/-/issues`)
        }
    }, [issue, props.history])

    const sidebarSections = useMemo<InfoSidebarSection[]>(
        () =>
            issue !== LOADING && issue !== null && !isErrorLike(issue)
                ? [
                      ...threadlikeSidebarSections({
                          thread: issue,
                          onThreadUpdate: onIssueUpdate,
                          extensionsController: props.extensionsController,
                      }),
                      {
                          expanded: (
                              <IssueDeleteButton
                                  {...props}
                                  issue={issue}
                                  buttonClassName="btn-link"
                                  className="btn-sm px-0 text-decoration-none"
                                  onDelete={onIssueDelete}
                              />
                          ),
                      },
                  ]
                : [],
        [issue, onIssueDelete, onIssueUpdate, props]
    )

    if (issue === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (issue === null) {
        return <HeroPage icon={AlertCircleIcon} title="Issue not found" />
    }
    if (isErrorLike(issue)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={issue.message} />
    }

    const context: IssueAreaContext = {
        ...props,
        issue,
        onIssueUpdate,
        setBreadcrumbItem,
    }

    return (
        <>
            <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: 100%; max-width: unset !important; overflow:hidden; padding: unset; } /* TODO!(sqs): hack */`}</style>
            <WithSidebar
                sidebarPosition="right"
                sidebar={<InfoSidebar sections={sidebarSections} />}
                className="flex-1"
            >
                <OverviewPagesArea<IssueAreaContext>
                    context={context}
                    header={header}
                    overviewComponent={IssueOverview}
                    pages={[
                        {
                            title: 'Activity',
                            path: '/',
                            exact: true,
                            render: () => <IssueActivity {...context} className={PAGE_CLASS_NAME} />,
                        },
                        {
                            title: 'Diagnostics',
                            path: '/diagnostics',
                            render: () => <IssueDiagnosticsList {...context} className={PAGE_CLASS_NAME} />,
                        },
                    ]}
                    location={props.location}
                    match={match}
                />
            </WithSidebar>
        </>
    )
}
