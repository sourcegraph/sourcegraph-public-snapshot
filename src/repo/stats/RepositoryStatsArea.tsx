import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { RepositoryStatsContributorsPage } from './RepositoryStatsContributorsPage'
import { RepositoryStatsNavbar } from './RepositoryStatsNavbar'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository stats page was not found."
    />
)

interface Props extends RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {
    repo: GQL.IRepository
}

/**
 * Properties passed to all page components in the repository stats area.
 */
export interface RepositoryStatsAreaPageProps {
    /**
     * The active repository.
     */
    repo: GQL.IRepository
}

const showNavbar = false

/**
 * Renders pages related to repository stats.
 */
export class RepositoryStatsArea extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const transferProps: RepositoryStatsAreaPageProps = {
            repo: this.props.repo,
        }

        return (
            <div className="repository-stats-area area--vertical">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="stats">Contributors</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                {showNavbar && (
                    <div className="area--vertical__navbar">
                        <RepositoryStatsNavbar className="area--vertical__navbar-inner" repo={this.props.repo.name} />
                    </div>
                )}
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        <Switch>
                            <Route
                                path={`${this.props.match.url}/contributors`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryStatsContributorsPage {...routeComponentProps} {...transferProps} />
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
