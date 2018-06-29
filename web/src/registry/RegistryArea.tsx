import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { platformEnabled } from '../user/tags'
import { RegistryAreaHeader } from './RegistryAreaHeader'
import { RegistryExtensionArea } from './RegistryExtensionArea'
import { RegistryNewExtensionPage } from './RegistryNewExtensionPage'
import { RegistryOverviewPage } from './RegistryOverviewPage'
import { RegistryPublisherPage } from './RegistryPublisherPage'

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

const AreaContent: React.SFC<{ children: JSX.Element }> = ({ children }) => (
    <div className="area--vertical__content">
        <div className="area--vertical__content-inner">{children}</div>
    </div>
)

interface Props extends RouteComponentProps<{ extensionID: string }>, ExtensionsProps, ExtensionsChangeProps {
    /**
     * The currently authenticated user.
     */
    user: GQL.IUser | null

    isLightTheme: boolean
}

/**
 * Properties passed to all page components in the registry area.
 */
export interface RegistryAreaPageProps extends ExtensionsProps, ExtensionsChangeProps {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * The registry area.
 */
export class RegistryArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        if (!this.props.user || !platformEnabled(this.props.user)) {
            return <NotFoundPage />
        }

        const transferProps: RegistryAreaPageProps = {
            authenticatedUser: this.props.user,
            extensions: this.props.extensions,
            onExtensionsChange: this.props.onExtensionsChange,
        }

        let showActions: 'primary' | 'link' | false
        if (this.props.location.pathname === '/registry') {
            showActions = 'primary'
        } else if (this.props.location.pathname === '/registry/extensions/new') {
            showActions = false
        } else {
            showActions = 'link'
        }

        return (
            <div className="registry-area area--vertical">
                <RegistryAreaHeader
                    className="area--vertical__header"
                    {...this.props}
                    {...transferProps}
                    showActions={showActions}
                    mainHeader={
                        this.props.location.pathname === this.props.match.path ||
                        this.props.location.pathname === `${this.props.match.path}/extensions/new`
                    }
                />
                <Switch>
                    <Route
                        path={`${this.props.match.url}`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <AreaContent>
                                <RegistryOverviewPage {...routeComponentProps} {...transferProps} />
                            </AreaContent>
                        )}
                    />
                    <Route
                        path={`${this.props.match.url}/extensions/new`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <RegistryNewExtensionPage {...routeComponentProps} {...transferProps} />
                        )}
                    />
                    {[
                        `${this.props.match.url}/extensions/:extensionID(.*)/-/`,
                        `${this.props.match.url}/extensions/:extensionID(.*)`,
                    ].map((path, i) => (
                        <Route
                            path={path}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RegistryExtensionArea
                                    {...routeComponentProps}
                                    {...transferProps}
                                    isLightTheme={this.props.isLightTheme}
                                />
                            )}
                        />
                    ))}
                    <Route
                        path={`${this.props.match.url}/publishers/:publisherType(users|organizations)/:publisherName`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <AreaContent>
                                <RegistryPublisherPage {...routeComponentProps} {...transferProps} />
                            </AreaContent>
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
