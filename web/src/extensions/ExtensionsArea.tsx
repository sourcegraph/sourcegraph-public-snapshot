import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { HeroPage } from '../components/HeroPage'
import { ThemeProps } from '../theme'
import { RouteDescriptor } from '../util/contributions'
import { ExtensionAreaRoute } from './extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extension/ExtensionAreaHeader'
import { ExtensionsAreaHeader, ExtensionsAreaHeaderActionButton } from './ExtensionsAreaHeader'

const NotFoundPage = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface ExtensionsAreaRoute extends RouteDescriptor<ExtensionsAreaRouteContext> {}

/**
 * Properties passed to all page components in the extensions area.
 */
export interface ExtensionsAreaRouteContext extends SettingsCascadeProps, PlatformContextProps, ThemeProps {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null

    /** The subject whose extensions and configuration to display. */
    subject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>
    extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute>
    extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
}

interface ExtensionsAreaProps
    extends RouteComponentProps<{ extensionID: string }>,
        SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps {
    routes: ReadonlyArray<ExtensionsAreaRoute>

    /**
     * The currently authenticated user.
     */
    authenticatedUser: GQL.IUser | null

    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>
    extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute>
    extensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton>
    extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
}

const LOADING: 'loading' = 'loading'

interface ExtensionsAreaState {}

/**
 * The extensions area.
 */
export class ExtensionsArea extends React.Component<ExtensionsAreaProps, ExtensionsAreaState> {
    public state: ExtensionsAreaState = {
        subjectOrError: LOADING,
    }

    private componentUpdates = new Subject<ExtensionsAreaProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: ExtensionsAreaProps): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const context: ExtensionsAreaRouteContext = {
            authenticatedUser: this.props.authenticatedUser,
            settingsCascade: this.props.settingsCascade,
            platformContext: this.props.platformContext,
            subject: this.props.viewerSubject,
            extensionAreaRoutes: this.props.extensionAreaRoutes,
            extensionAreaHeaderNavItems: this.props.extensionAreaHeaderNavItems,
            isLightTheme: this.props.isLightTheme,
        }

        return (
            <div className="extensions-area area--vertical">
                <ExtensionsAreaHeader
                    {...this.props}
                    {...context}
                    actionButtons={this.props.extensionsAreaHeaderActionButtons}
                    isPrimaryHeader={this.props.location.pathname === this.props.match.path}
                />
                <Switch>
                    {this.props.routes.map(
                        ({ path, exact, condition = () => true, render }) =>
                            condition(context) && (
                                <Route
                                    key="hardcoded-key"
                                    path={this.props.match.url + path}
                                    exact={exact}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                />
                            )
                    )}
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
