import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, graphQLContent } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { ConfigurationCascadeProps, ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { RouteDescriptor } from '../../util/contributions'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { ExtensionsAreaRouteContext } from '../ExtensionsArea'
import { ExtensionAreaHeader, ExtensionAreaHeaderNavItem } from './ExtensionAreaHeader'

export const registryExtensionFragment = gql`
    fragment RegistryExtensionFields on RegistryExtension {
        id
        publisher {
            __typename
            ... on User {
                id
                username
                displayName
                url
            }
            ... on Org {
                id
                name
                displayName
                url
            }
        }
        extensionID
        extensionIDWithoutRegistry
        name
        manifest {
            raw
            title
            description
        }
        createdAt
        updatedAt
        url
        remoteURL
        registryName
        isLocal
        viewerCanAdminister
    }
`

const NotFoundPage = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface ExtensionAreaRoute extends RouteDescriptor<ExtensionAreaRouteContext> {}

export interface ExtensionAreaProps extends ExtensionsAreaRouteContext, RouteComponentProps<{ extensionID: string }> {
    routes: ReadonlyArray<ExtensionAreaRoute>
    isLightTheme: boolean
    extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
}

interface ExtensionAreaState {
    /** The registry extension, undefined while loading, or an error.  */
    extensionOrError?: ConfiguredExtension<GQL.IRegistryExtension> | ErrorLike
}

/**
 * Properties passed to all page components in the registry extension area.
 */
export interface ExtensionAreaRouteContext extends ConfigurationCascadeProps, ExtensionsProps {
    /** The extension registry area main URL. */
    url: string

    /** The extension that is the subject of the page. */
    extension: ConfiguredExtension<GQL.IRegistryExtension>

    /** Called when the component updates the extension and it should be refreshed here. */
    onDidUpdateExtension: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

/**
 * An extension area.
 */
export class ExtensionArea extends React.Component<ExtensionAreaProps> {
    public state: ExtensionAreaState = {}

    private componentUpdates = new Subject<ExtensionAreaProps>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const routeMatchChanges = this.componentUpdates.pipe(
            map(({ match }) => match.params),
            distinctUntilChanged()
        )

        // Changes to the route-matched extension ID.
        const extensionIDChanges = routeMatchChanges.pipe(
            map(({ extensionID }) => extensionID),
            distinctUntilChanged()
        )

        // Changes to the global extensions settings.
        const globalExtensionsSettingsChanges = this.componentUpdates.pipe(
            map(({ extensions }) => extensions),
            distinctUntilChanged()
        )

        // Fetch extension.
        this.subscriptions.add(
            combineLatest(
                extensionIDChanges,
                merge(
                    this.refreshRequests.pipe(mapTo(false)),
                    globalExtensionsSettingsChanges.pipe(mapTo(false)),
                    of(false)
                )
            )
                .pipe(
                    switchMap(([extensionID, forceRefresh]) => {
                        type PartialStateUpdate = Pick<ExtensionAreaState, 'extensionOrError'>
                        return this.props.extensions
                            .forExtensionID(extensionID, registryExtensionFragment[graphQLContent])
                            .pipe(
                                catchError(error => [error]),
                                map(c => ({ extensionOrError: c } as PartialStateUpdate)),

                                // Don't clear old data while we reload, to avoid unmounting all components during
                                // loading.
                                startWith<PartialStateUpdate>(forceRefresh ? { extensionOrError: undefined } : {})
                            )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: ExtensionAreaProps): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.extensionOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.extensionOrError)) {
            return (
                <HeroPage
                    icon={AlertCircleIcon}
                    title="Error"
                    subtitle={upperFirst(this.state.extensionOrError.message)}
                />
            )
        }

        // The URL, without the trailing "/-" that `this.props.match.url` includes on sub-pages.
        const url = this.props.match.url.replace(/\/-\/?$/, '')

        const context: ExtensionAreaRouteContext = {
            url,
            authenticatedUser: this.props.authenticatedUser,
            onDidUpdateExtension: this.onDidUpdateExtension,
            configurationCascade: this.props.configurationCascade,
            extension: this.state.extensionOrError,
            extensions: this.props.extensions,
            isLightTheme: this.props.isLightTheme,
        }

        return (
            <div className="registry-extension-area area--vertical">
                <ExtensionAreaHeader {...this.props} {...context} navItems={this.props.extensionAreaHeaderNavItems} />
                <div className="container pt-3">
                    <Switch>
                        {this.props.routes.map(
                            ({ path, render, exact, condition = () => true }) =>
                                condition(context) && (
                                    <Route
                                        path={url + path}
                                        exact={exact}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        // tslint:disable-next-line:jsx-no-lambda
                                        render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                    />
                                )
                        )}
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }

    private onDidUpdateExtension = () => this.refreshRequests.next()
}
