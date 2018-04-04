import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../../components/HeroPage'
import { DismissibleAlert } from '../../site/DismissibleAlert'
import { RepoHeaderActionPortal } from '../RepoHeaderActionPortal'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepositoryBranchesAllPage } from './RepositoryBranchesAllPage'
import { RepositoryBranchesNavbar } from './RepositoryBranchesNavbar'
import { RepositoryBranchesOverviewPage } from './RepositoryBranchesOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository branches page was not found."
    />
)

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
}

interface State {
    error?: string
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
    public state: State = {}

    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        const transferProps: { repo: GQL.IRepository } = {
            repo: this.props.repo,
        }

        return (
            <div className="repository-graph-area area--vertical">
                <RepoHeaderActionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="branches">Branches</RepoHeaderBreadcrumbNavItem>}
                />
                <div>
                    <DismissibleAlert
                        className="alert-warning mb-1"
                        partialStorageKey="repository-branches-experimental"
                    >
                        <span>
                            The repository branches area is an <strong>experimental</strong> feature.
                        </span>
                    </DismissibleAlert>
                </div>
                <RepositoryBranchesNavbar className="area--vertical__navbar" repo={this.props.repo.uri} />
                <div className="area--vertical__content">
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
        )
    }
}
