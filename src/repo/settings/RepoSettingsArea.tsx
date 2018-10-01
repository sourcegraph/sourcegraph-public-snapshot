import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import DoNotDisturbIcon from 'mdi-react/DoNotDisturbIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { fetchRepository } from './backend'
import { RepoSettingsIndexPage } from './RepoSettingsIndexPage'
import { RepoSettingsMirrorPage } from './RepoSettingsMirrorPage'
import { RepoSettingsOptionsPage } from './RepoSettingsOptionsPage'
import { RepoSettingsSidebar } from './RepoSettingsSidebar'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository page was not found."
    />
)

interface Props extends RouteComponentProps<any>, RepoHeaderContributionsLifecycleProps {
    repo: GQL.IRepository
    user: GQL.IUser | null
    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
}

interface State {
    repo?: GQL.IRepository | null
    error?: string
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * a repository's settings.
 */
export class RepoSettingsArea extends React.Component<Props> {
    public state: State = {}

    private repoChanges = new Subject<GQL.IRepository>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.repoChanges
                .pipe(switchMap(({ name }) => fetchRepository(name)))
                .subscribe(repo => this.setState({ repo }), err => this.setState({ error: err.message }))
        )
        this.repoChanges.next(this.props.repo)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repo !== this.props.repo) {
            this.repoChanges.next(props.repo)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        if (this.state.repo === undefined) {
            return null
        }
        if (this.state.repo === null) {
            return <NotFoundPage />
        }
        if (!this.state.repo.viewerCanAdminister) {
            return (
                <HeroPage
                    icon={DoNotDisturbIcon}
                    title="Forbidden"
                    subtitle="You are not authorized to view or change this repository's settings."
                />
            )
        }
        if (!this.props.user) {
            return null
        }

        const transferProps: { user: GQL.IUser; repo: GQL.IRepository } = {
            user: this.props.user,
            repo: this.state.repo,
        }

        return (
            <div className="repo-settings-area area">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={
                        <span key="graph" className="repo-settings-area__header-item">
                            Settings
                        </span>
                    }
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <RepoSettingsSidebar className="area__sidebar" {...this.props} {...transferProps} />
                <div className="area__content">
                    <Switch>
                        <Route
                            path={`${this.props.match.url}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepoSettingsOptionsPage
                                    {...routeComponentProps}
                                    {...transferProps}
                                    onDidUpdateRepository={this.props.onDidUpdateRepository}
                                />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/index`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepoSettingsIndexPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/mirror`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepoSettingsMirrorPage
                                    {...routeComponentProps}
                                    {...transferProps}
                                    onDidUpdateRepository={this.props.onDidUpdateRepository}
                                />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
