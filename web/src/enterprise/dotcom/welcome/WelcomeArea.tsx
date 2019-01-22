import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { RouteDescriptor } from '../../../util/contributions'
import { WelcomeAreaFooter } from './WelcomeAreaFooter'

const NotFoundPage = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface WelcomeAreaRoute extends RouteDescriptor<WelcomeAreaRouteContext> {}

export interface WelcomeAreaProps extends ExtensionsControllerProps, PlatformContextProps, RouteComponentProps<{}> {
    isLightTheme: boolean
    routes: ReadonlyArray<WelcomeAreaRoute>
    location: H.Location
    history: H.History
}

export interface WelcomeAreaRouteContext extends WelcomeAreaProps {}

/**
 * The welcome area, which contains general product information.
 */
export class WelcomeArea extends React.PureComponent<WelcomeAreaProps> {
    public render(): JSX.Element | null {
        const { children, ...context } = this.props

        return (
            <div className="welcome-area container">
                {this.props.location.pathname !== '/welcome' && (
                    <nav className="d-block my-2">
                        <Link to="/welcome" className="py-2 pr-2">
                            <ChevronLeftIcon className="icon-inline" />
                            Welcome
                        </Link>
                    </nav>
                )}
                <ErrorBoundary location={this.props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="icon-inline my-2 d-block mx-auto" />}>
                        <Switch>
                            {this.props.routes.map(
                                ({ path, exact, render, condition = () => true }) =>
                                    condition(context) && (
                                        <Route
                                            path={this.props.match.url + path}
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            exact={exact}
                                            // tslint:disable-next-line:jsx-no-lambda
                                            render={routeComponentProps => (
                                                <>
                                                    {render({ ...context, ...routeComponentProps })}
                                                    <WelcomeAreaFooter isLightTheme={this.props.isLightTheme} />
                                                </>
                                            )}
                                        />
                                    )
                            )}
                            <Route component={NotFoundPage} key="hardcoded-key" />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        )
    }
}
