import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { USE_PLATFORM } from '../extensions/environment/ExtensionsEnvironment'
import { ExtensionArea } from './extension/ExtensionArea'
import { ExtensionsAreaHeader } from './ExtensionsAreaHeader'
import { ConfigurationCascadeProps, ExtensionsProps } from './ExtensionsClientCommonContext'
import { ExtensionsOverviewPage } from './ExtensionsOverviewPage'
import { RegistryArea } from './registry/RegistryArea'

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

interface Props extends RouteComponentProps<{ extensionID: string }>, ConfigurationCascadeProps, ExtensionsProps {
    /**
     * The currently authenticated user.
     */
    user: GQL.IUser | null

    viewerSubject: Pick<GQL.IConfigurationSubject, 'id' | 'viewerCanAdminister'>

    isLightTheme: boolean
}

/**
 * Properties passed to all page components in the extensions area.
 */
export interface ExtensionsAreaPageProps extends ConfigurationCascadeProps, ExtensionsProps {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null

    /** The subject whose extensions and configuration to display. */
    subject: Pick<GQL.IConfigurationSubject, 'id' | 'viewerCanAdminister'>
}

const LOADING: 'loading' = 'loading'

interface State {}

/**
 * The extensions area.
 */
export class ExtensionsArea extends React.Component<Props, State> {
    public state: State = {
        subjectOrError: LOADING,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: Props): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!USE_PLATFORM) {
            return <NotFoundPage />
        }

        const transferProps: ExtensionsAreaPageProps = {
            authenticatedUser: this.props.user,
            configurationCascade: this.props.configurationCascade,
            extensions: this.props.extensions,
            subject: this.props.viewerSubject,
        }

        return (
            <div className="extensions-area area--vertical">
                <ExtensionsAreaHeader
                    {...this.props}
                    {...transferProps}
                    isPrimaryHeader={this.props.location.pathname === this.props.match.path}
                />
                <Switch>
                    <Route
                        path={this.props.match.url}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <ExtensionsOverviewPage {...routeComponentProps} {...transferProps} />
                        )}
                    />
                    <Route
                        path={`${this.props.match.url}/registry`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <RegistryArea {...routeComponentProps} {...transferProps} />}
                    />
                    {[`${this.props.match.url}/:extensionID(.*)/-/`, `${this.props.match.url}/:extensionID(.*)`].map(
                        path => (
                            <Route
                                path={path}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <ExtensionArea
                                        {...routeComponentProps}
                                        {...transferProps}
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                )}
                            />
                        )
                    )}
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
