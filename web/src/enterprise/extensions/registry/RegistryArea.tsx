import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { ExtensionsAreaRouteContext } from '../../../extensions/ExtensionsArea'
import { RegistryNewExtensionPage } from './RegistryNewExtensionPage'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

interface Props extends RouteComponentProps<{}>, ExtensionsAreaRouteContext {}

/**
 * Properties passed to all page components in the registry area.
 */
export interface RegistryAreaPageProps extends PlatformContextProps {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * The extension registry area.
 */
export class RegistryArea extends React.Component<Props> {
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
        const transferProps: RegistryAreaPageProps = {
            authenticatedUser: this.props.authenticatedUser,
            platformContext: this.props.platformContext,
        }

        return (
            <div className="registry-area area--vertical">
                <Switch>
                    <Route
                        path={`${this.props.match.url}/new`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <RegistryNewExtensionPage {...routeComponentProps} {...transferProps} />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
