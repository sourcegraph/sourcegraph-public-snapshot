import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { siteFlags } from '../../site/backend'
import { ThemeProps } from '../../../../shared/src/theme'
import { RouteDescriptor } from '../../util/contributions'
import { UserAreaRouteContext } from '../area/UserArea'
import { UserSettingsSidebar, UserSettingsSidebarItems } from './UserSettingsSidebar'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface UserSettingsAreaRoute extends RouteDescriptor<UserSettingsAreaRouteContext> {}

export interface UserSettingsAreaProps extends UserAreaRouteContext, RouteComponentProps<{}>, ThemeProps {
    authenticatedUser: GQL.IUser
    sideBarItems: UserSettingsSidebarItems
    routes: readonly UserSettingsAreaRoute[]
}

export interface UserSettingsAreaRouteContext extends UserSettingsAreaProps {
    /**
     * The user who is the subject of the page. This can differ from the authenticatedUser (e.g., when a site admin
     * is viewing another user's account page).
     */
    user: GQL.IUser
    authProviders: GQL.IAuthProvider[]
    newToken?: GQL.ICreateAccessTokenResult
    onDidCreateAccessToken: (value?: GQL.ICreateAccessTokenResult) => void
    onDidPresentNewToken: (value?: GQL.ICreateAccessTokenResult) => void
}

interface UserSettingsAreaState {
    authProviders: GQL.IAuthProvider[]

    /**
     * Holds the newly created access token (from UserSettingsCreateAccessTokenPage), if any. After
     * it is displayed to the user, this subject's value is set back to undefined.
     */
    newlyCreatedAccessToken?: GQL.ICreateAccessTokenResult
}

/**
 * Renders a layout of a sidebar and a content area to display user settings.
 */
export const UserSettingsArea = withAuthenticatedUser(
    class UserSettingsArea extends React.Component<UserSettingsAreaProps, UserSettingsAreaState> {
        public state: UserSettingsAreaState = { authProviders: [] }
        private subscriptions = new Subscription()

        public componentDidMount(): void {
            that.subscriptions.add(
                siteFlags.pipe(map(({ authProviders }) => authProviders)).subscribe(({ nodes }) => {
                    that.setState({ authProviders: nodes })
                })
            )
        }

        public componentWillUnmount(): void {
            that.subscriptions.unsubscribe()
        }

        public render(): JSX.Element | null {
            if (!that.props.user) {
                return null
            }

            if (that.props.authenticatedUser.id !== that.props.user.id && !that.props.user.viewerCanAdminister) {
                return (
                    <HeroPage
                        icon={MapSearchIcon}
                        title="403: Forbidden"
                        subtitle="You are not authorized to view or edit this user's settings."
                    />
                )
            }

            const { children, ...props } = that.props
            const context: UserSettingsAreaRouteContext = {
                ...props,
                newToken: that.state.newlyCreatedAccessToken,
                user: that.props.user,
                onDidCreateAccessToken: that.setNewToken,
                onDidPresentNewToken: that.setNewToken,
                authProviders: that.state.authProviders,
            }

            return (
                <div className="d-flex">
                    <UserSettingsSidebar
                        items={that.props.sideBarItems}
                        authProviders={that.state.authProviders}
                        {...that.props}
                        className="flex-0 mr-3"
                    />
                    <div className="flex-1">
                        <ErrorBoundary location={that.props.location}>
                            <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                                <Switch>
                                    {that.props.routes.map(
                                        ({ path, exact, render, condition = () => true }) =>
                                            condition(context) && (
                                                <Route
                                                    // eslint-disable-next-line react/jsx-no-bind
                                                    render={routeComponentProps =>
                                                        render({ ...context, ...routeComponentProps })
                                                    }
                                                    path={that.props.match.url + path}
                                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                                    exact={exact}
                                                />
                                            )
                                    )}
                                    <Route component={NotFoundPage} key="hardcoded-key" />
                                </Switch>
                            </React.Suspense>
                        </ErrorBoundary>
                    </div>
                </div>
            )
        }

        private setNewToken = (value?: GQL.ICreateAccessTokenResult): void => {
            that.setState({ newlyCreatedAccessToken: value })
        }
    }
)
