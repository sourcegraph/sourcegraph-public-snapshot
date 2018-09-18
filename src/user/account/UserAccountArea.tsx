import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { siteFlags } from '../../site/backend'
import { RouteDescriptor } from '../../util/contributions'
import { UserAreaPageProps } from '../area/UserArea'
import { UserAccountSidebar, UserAccountSidebarItems } from './UserAccountSidebar'

const NotFoundPage = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface UserAccountAreaRoute extends RouteDescriptor<UserAccountAreaRouteContext> {}

export interface UserAccountAreaProps extends UserAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
    sideBarItems: UserAccountSidebarItems
    routes: ReadonlyArray<UserAccountAreaRoute>
}

export interface UserAccountAreaRouteContext extends UserAccountAreaProps {
    user: GQL.IUser
    externalAuthEnabled: boolean
    newToken?: GQL.ICreateAccessTokenResult
    onDidCreateAccessToken: (value?: GQL.ICreateAccessTokenResult) => void
    onDidPresentNewToken: (value?: GQL.ICreateAccessTokenResult) => void
}

interface UserAccountAreaState {
    externalAuthEnabled: boolean

    /**
     * Holds the newly created access token (from UserAccountCreateAccessTokenPage), if any. After
     * it is displayed to the user, this subject's value is set back to undefined.
     */
    newlyCreatedAccessToken?: GQL.ICreateAccessTokenResult
}

/**
 * Renders a layout of a sidebar and a content area to display user settings.
 */
export class UserAccountArea extends React.Component<UserAccountAreaProps, UserAccountAreaState> {
    public state: UserAccountAreaState = { externalAuthEnabled: false }
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            siteFlags.pipe(map(({ externalAuthEnabled }) => externalAuthEnabled)).subscribe(externalAuthEnabled => {
                this.setState({ externalAuthEnabled })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.props.user) {
            return null
        }

        // If not logged in, redirect to sign in
        if (!this.props.authenticatedUser) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        if (this.props.authenticatedUser.id !== this.props.user.id && !this.props.user.viewerCanAdminister) {
            return (
                <HeroPage
                    icon={MapSearchIcon}
                    title="403: Forbidden"
                    subtitle="You are not authorized to view or edit this user's settings."
                />
            )
        }

        if (this.props.match.isExact) {
            return <Redirect to={`${this.props.match.path}/profile`} />
        }

        const { children, ...props } = this.props
        const context: UserAccountAreaRouteContext = {
            ...props,
            newToken: this.state.newlyCreatedAccessToken,
            user: this.props.user,
            onDidCreateAccessToken: this.setNewToken,
            onDidPresentNewToken: this.setNewToken,
            externalAuthEnabled: this.state.externalAuthEnabled,
        }

        return (
            <div className="user-settings-area area">
                <UserAccountSidebar
                    externalAuthEnabled={this.state.externalAuthEnabled}
                    items={this.props.sideBarItems}
                    {...this.props}
                    className="area__sidebar"
                />
                <div className="area__content">
                    <Switch>
                        {this.props.routes.map(
                            ({ path, exact, render, condition = () => true }) =>
                                condition(context) && (
                                    <Route
                                        path={this.props.match.url + path}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        exact={exact}
                                        // tslint:disable-next-line:jsx-no-lambda
                                        render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                    />
                                )
                        )}
                        <Route component={NotFoundPage} key="hardcoded-key" />
                    </Switch>
                </div>
            </div>
        )
    }

    private setNewToken = (value?: GQL.ICreateAccessTokenResult): void => {
        this.setState({ newlyCreatedAccessToken: value })
    }
}
