import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isDefined } from '../../../../../shared/src/util/types'
import { BreadcrumbItem, Breadcrumbs } from '../../../components/breadcrumbs/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { RepoContainerContext } from '../../../repo/RepoContainer'
import { RepoHeaderBreadcrumbNavItem } from '../../../repo/RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../../../repo/RepoHeaderContributionPortal'
import { ThemeProps } from '../../../theme'
import { IssueArea } from '../detail/IssueArea'
import { RepositoryIssuesListPage } from './list/RepositoryIssuesListPage'
import { IssuesNewPage } from './new/IssuesNewPage'

export interface RepositoryIssuesAreaContext
    extends Pick<RepoContainerContext, 'repo' | 'repoHeaderContributionsLifecycleProps'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The URL to the repository issues area. */
    issuesURL: string

    setBreadcrumbItem?: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    authenticatedUser: GQL.IUser | null
}

interface Props
    extends Pick<
        RepositoryIssuesAreaContext,
        Exclude<keyof RepositoryIssuesAreaContext, 'issuesURL' | 'setBreadcrumbItem'>
    > {}

/**
 * The issues area for a repository.
 */
export const RepositoryIssuesArea: React.FunctionComponent<Props> = ({
    repoHeaderContributionsLifecycleProps,
    ...props
}) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: RepositoryIssuesAreaContext = {
        ...props,
        issuesURL: `${props.repo.url}/-/issues`,
        setBreadcrumbItem,
    }
    const newIssueURL = `${context.issuesURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: displayRepoName(props.repo.name), to: props.repo.url },
                { text: 'Issues', to: context.issuesURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.issuesURL, props.repo.name, props.repo.url]
    )

    const breadcrumbs = <Breadcrumbs items={breadcrumbItems} className="my-4" />

    return (
        <>
            <style>{`.repo-header{display:none !important;}` /* TODO!(sqs) */}</style>
            <RepoHeaderContributionPortal
                position="nav"
                element={<RepoHeaderBreadcrumbNavItem key="issues">Issues</RepoHeaderBreadcrumbNavItem>}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
            <Switch>
                <Route path={context.issuesURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <RepositoryIssuesListPage {...context} newIssueURL={newIssueURL} />
                    </div>
                </Route>
                <Route path={newIssueURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <IssuesNewPage {...context} />
                    </div>
                </Route>
                <Route
                    path={`${context.issuesURL}/:issueNumber`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ issueNumber: string }>) => (
                        <IssueArea
                            {...context}
                            {...routeComponentProps}
                            issueNumber={routeComponentProps.match.params.issueNumber}
                            header={breadcrumbs}
                        />
                    )}
                />
                <Route>
                    <HeroPage
                        icon={MapSearchIcon}
                        title="404: Not Found"
                        subtitle="Sorry, the requested page was not found."
                    />
                </Route>
            </Switch>
        </>
    )
}
