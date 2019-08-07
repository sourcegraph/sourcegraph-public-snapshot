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
import { ChangesetArea } from '../detail/ChangesetArea'
import { RepositoryChangesetsListPage } from './list/RepositoryChangesetsListPage'
import { ChangesetsNewPage } from './new/ChangesetsNewPage'

export interface RepositoryChangesetsAreaContext
    extends Pick<RepoContainerContext, 'repo' | 'repoHeaderContributionsLifecycleProps'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The URL to the repository changesets area. */
    changesetsURL: string

    setBreadcrumbItem?: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    authenticatedUser: GQL.IUser | null
}

interface Props
    extends Pick<
        RepositoryChangesetsAreaContext,
        Exclude<keyof RepositoryChangesetsAreaContext, 'changesetsURL' | 'setBreadcrumbItem'>
    > {}

/**
 * The changesets area for a repository.
 */
export const RepositoryChangesetsArea: React.FunctionComponent<Props> = ({
    repoHeaderContributionsLifecycleProps,
    ...props
}) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: RepositoryChangesetsAreaContext = {
        ...props,
        changesetsURL: `${props.repo.url}/-/changesets`,
        setBreadcrumbItem,
    }
    const newChangesetURL = `${context.changesetsURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: displayRepoName(props.repo.name), to: props.repo.url },
                { text: 'Changesets', to: context.changesetsURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.changesetsURL, props.repo.name, props.repo.url]
    )

    const breadcrumbs = <Breadcrumbs items={breadcrumbItems} className="my-4" />

    return (
        <>
            <style>{`.repo-header{display:none !important;}` /* TODO!(sqs) */}</style>
            <RepoHeaderContributionPortal
                position="nav"
                element={<RepoHeaderBreadcrumbNavItem key="changesets">Changesets</RepoHeaderBreadcrumbNavItem>}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
            <Switch>
                <Route path={context.changesetsURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <RepositoryChangesetsListPage {...context} newChangesetURL={newChangesetURL} />
                    </div>
                </Route>
                <Route path={newChangesetURL} exact={true}>
                    <div className="container">
                        {breadcrumbs}
                        <ChangesetsNewPage {...context} />
                    </div>
                </Route>
                <Route
                    path={`${context.changesetsURL}/:changesetNumber`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ changesetNumber: string }>) => (
                        <ChangesetArea
                            {...context}
                            {...routeComponentProps}
                            changesetNumber={routeComponentProps.match.params.changesetNumber}
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
