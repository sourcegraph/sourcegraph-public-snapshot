import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
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

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
    user: GQL.IUser | null
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
                .pipe(switchMap(({ uri }) => fetchRepository(uri)))
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
            return <HeroPage icon={DirectionalSignIcon} title="Error" subtitle={this.state.error} />
        }

        if (this.state.repo === undefined || !this.props.user) {
            return null
        }
        if (this.state.repo === null) {
            return <NotFoundPage />
        }
        if (!this.state.repo.viewerCanAdminister) {
            return <HeroPage icon={DirectionalSignIcon} title="Repository administrators only" />
        }

        const transferProps: { user: GQL.IUser; repo: GQL.IRepository } = {
            user: this.props.user,
            repo: this.state.repo,
        }

        return (
            <div className="repo-settings-area area">
                <RepoSettingsSidebar {...this.props} {...transferProps} />
                <div className="area__content">
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
