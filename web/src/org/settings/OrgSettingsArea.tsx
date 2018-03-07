import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'
import { HeroPage } from '../../components/HeroPage'
import { fetchOrg } from '../backend'
import { OrgSettingsConfigurationPage } from './OrgSettingsConfigurationPage'
import { OrgSettingsMembersPage } from './OrgSettingsMembersPage'
import { OrgSettingsProfilePage } from './OrgSettingsProfilePage'
import { OrgSettingsSidebar } from './OrgSettingsSidebar'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

interface Props extends RouteComponentProps<{ orgName: string }> {
    isLightTheme: boolean
}

interface State {
    org?: GQL.IOrg | null
    user?: GQL.IUser | null
    error?: string
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export class OrgSettingsArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ orgName: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.routeMatchChanges
                .pipe(
                    withLatestFrom(currentUser),
                    mergeMap(([{ orgName }, user]) => {
                        const org = user!.orgs.find(org => org.name === orgName)
                        if (!org) {
                            throw new Error('org not found')
                        }
                        return fetchOrg(org.id)
                    })
                )
                .subscribe(org => this.setState({ org }), err => this.setState({ error: err.message }))
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
        if (
            this.props.location.pathname === this.props.match.url ||
            this.props.location.pathname === `${this.props.match.url}/settings`
        ) {
            return <Redirect to={`${this.props.match.url}/settings/profile`} />
        }

        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        if (this.state.org === undefined || !this.state.user) {
            return null
        }
        if (this.state.org === null) {
            return <NotFoundPage />
        }

        const transferProps = { user: this.state.user, org: this.state.org, isLightTheme: this.props.isLightTheme }

        return (
            <div className="org-settings-area area">
                <OrgSettingsSidebar className="area__sidebar" {...this.props} />
                <div className="area__content">
                    <Switch>
                        <Route
                            path={`${this.props.match.url}/settings/profile`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <OrgSettingsProfilePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/settings/members`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <OrgSettingsMembersPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/settings/configuration`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <OrgSettingsConfigurationPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
