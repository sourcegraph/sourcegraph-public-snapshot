import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../../../../src/backend/graphqlschema'
import { DismissibleAlert } from '../../../../src/components/DismissibleAlert'
import { HeroPage } from '../../../../src/components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../../../../src/repo/RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../../../../src/repo/RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../../../../src/repo/RepoHeaderContributionPortal'
import { RepositoryGraphDependenciesPage } from './RepositoryGraphDependenciesPage'
import { RepositoryGraphOverviewPage } from './RepositoryGraphOverviewPage'
import { RepositoryGraphPackagesPage } from './RepositoryGraphPackagesPage'
import { RepositoryGraphSidebar } from './RepositoryGraphSidebar'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository graph page was not found."
    />
)

interface Props extends RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {
    repo: GQL.IRepository
    rev: string | undefined
    commitID: string
    defaultBranch: string
    routePrefix: string
}

interface State {
    error?: string
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * the repository graph.
 */
export class RepositoryGraphArea extends React.Component<Props> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        const transferProps: {
            repo: GQL.IRepository
            rev: string | undefined
            commitID: string
        } = {
            repo: this.props.repo,
            rev: this.props.rev,
            commitID: this.props.commitID,
        }

        return (
            <div className="repository-graph-area area">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="graph">Graph</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <RepositoryGraphSidebar
                    className="area__sidebar"
                    {...this.props}
                    {...transferProps}
                    routePrefix={this.props.routePrefix}
                />
                <div className="area__content">
                    <DismissibleAlert className="alert-warning mb-4" partialStorageKey="repository-graph-experimental">
                        <span>
                            The repository graph area is an <strong>experimental</strong> feature that lets you explore
                            a repository's dependencies and packages. Not all languages and repositories are supported.
                        </span>
                    </DismissibleAlert>
                    <Switch>
                        <Route
                            path={`${this.props.match.url}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGraphOverviewPage
                                    {...routeComponentProps}
                                    {...transferProps}
                                    routePrefix={this.props.routePrefix}
                                />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/packages`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGraphPackagesPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/dependencies`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGraphDependenciesPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
