import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { RepositoryBranchesAllPage } from './RepositoryBranchesAllPage'
import { RepositoryBranchesNavbar } from './RepositoryBranchesNavbar'
import { RepositoryBranchesOverviewPage } from './RepositoryBranchesOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository branches page was not found."
    />
)

interface Props extends RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {
    repo: GQL.IRepository
}

/**
 * Properties passed to all page components in the repository branches area.
 */
export interface RepositoryBranchesAreaPageProps {
    /**
     * The active repository.
     */
    repo: GQL.IRepository
}

/**
 * Renders pages related to repository branches.
 */
export class RepositoryBranchesArea extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const transferProps: { repo: GQL.IRepository } = {
            repo: this.props.repo,
        }

        return (
            <div className="repository-branches-area area--vertical">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="branches">Branches</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <div className="area--vertical__navbar">
                    <RepositoryBranchesNavbar className="area--vertical__navbar-inner" repo={this.props.repo.name} />
                </div>
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        <Switch>
                            <Route
                                path={`${this.props.match.url}`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryBranchesOverviewPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route
                                path={`${this.props.match.url}/all`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryBranchesAllPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route key="hardcoded-key" component={NotFoundPage} />
                        </Switch>
                    </div>
                </div>
            </div>
        )
    }
}
