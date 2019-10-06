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
import { ThreadArea } from '../detail/ThreadArea'
import { RepositoryThreadsListPage } from './list/RepositoryThreadsListPage'
import { ThreadsNewPage } from './new/ThreadsNewPage'

export interface RepositoryThreadsAreaContext
    extends Pick<RepoContainerContext, 'repo' | 'repoHeaderContributionsLifecycleProps'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The URL to the repository threads area. */
    threadsURL: string

    setBreadcrumbItem?: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    authenticatedUser: GQL.IUser | null
}

interface Props
    extends Pick<
        RepositoryThreadsAreaContext,
        Exclude<keyof RepositoryThreadsAreaContext, 'threadsURL' | 'setBreadcrumbItem'>
    > {}

/**
 * The threads area for a repository.
 */
export const RepositoryThreadsArea: React.FunctionComponent<Props> = ({
    repoHeaderContributionsLifecycleProps,
    ...props
}) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: RepositoryThreadsAreaContext = {
        ...props,
        threadsURL: `${props.repo.url}/-/threads`,
        setBreadcrumbItem,
    }
    const newThreadURL = `${context.threadsURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: displayRepoName(props.repo.name), to: props.repo.url },
                { text: 'Threads', to: context.threadsURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.threadsURL, props.repo.name, props.repo.url]
    )

    const breadcrumbs = <Breadcrumbs items={breadcrumbItems} className="my-4" />

    return (
        <>
            <style>{`.repo-header{display:none !important;}` /* TODO!(sqs) */}</style>
            <RepoHeaderContributionPortal
                position="nav"
                element={<RepoHeaderBreadcrumbNavItem key="threads">Threads</RepoHeaderBreadcrumbNavItem>}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
            <Switch>
                <Route path={context.threadsURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <RepositoryThreadsListPage {...context} newThreadURL={newThreadURL} />
                    </div>
                </Route>
                <Route path={newThreadURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <ThreadsNewPage {...context} />
                    </div>
                </Route>
                <Route
                    path={`${context.threadsURL}/:threadNumber`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ threadNumber: string }>) => (
                        <ThreadArea
                            {...context}
                            {...routeComponentProps}
                            threadNumber={routeComponentProps.match.params.threadNumber}
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
