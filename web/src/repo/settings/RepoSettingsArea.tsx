import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'
import { RepoBreadcrumb } from '../../components/Breadcrumb'
import { HeroPage } from '../../components/HeroPage'
import { RouteWithProps } from '../../util/RouteWithProps'
import { fetchRepository } from './backend'
import { RepoSettingsOptionsPage } from './RepoSettingsOptionsPage'
import { RepoSettingsSidebar } from './RepoSettingsSidebar'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository page was not found."
    />
)

interface Props extends RouteComponentProps<{ repo: string }> {}

interface State {
    repo?: GQL.IRepository | null
    user?: GQL.IUser | null
    error?: string
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * a repository's settings.
 */
export class RepoSettingsArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ repo: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.routeMatchChanges
                .pipe(mergeMap(({ repo }) => fetchRepository(repo)))
                .subscribe(repo => this.setState({ repo }), err => this.setState({ error: err.message }))
        )
        this.routeMatchChanges.next(this.props.match.params)

        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.match.params !== this.props.match.params) {
            this.routeMatchChanges.next(props.match.params)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={DirectionalSignIcon} title="Error" subtitle={this.state.error} />
        }

        if (this.state.repo === undefined || !this.state.user) {
            return null
        }
        if (this.state.repo === null) {
            return <NotFoundPage />
        }
        if (!this.state.repo.viewerCanAdminister) {
            return <HeroPage icon={DirectionalSignIcon} title="Repository administrators only" />
        }

        const transferProps: { user: GQL.IUser; repo: GQL.IRepository } = {
            user: this.state.user,
            repo: this.state.repo,
        }

        return (
            <div className="repo-settings-area area">
                <RepoSettingsSidebar {...this.props} {...transferProps} />
                <div className="area__content">
                    <div>
                        <Link
                            to={`/${this.state.repo.uri}`}
                            className="sidebar__action-button btn area__content-header"
                        >
                            <RepoIcon className="icon-inline sidebar__action-icon" />
                            <RepoBreadcrumb repoPath={this.state.repo.uri} disableLinks={true} />
                        </Link>
                    </div>
                    <Switch>
                        <RouteWithProps
                            path={`${this.props.match.url}`}
                            component={RepoSettingsOptionsPage}
                            exact={true}
                            other={transferProps}
                        />
                        <Route component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
