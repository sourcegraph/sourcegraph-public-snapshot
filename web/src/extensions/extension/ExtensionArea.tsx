import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, graphQLContent } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { ConfigurationCascadeProps, ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { RegistryExtensionManifestPage } from '../extension/RegistryExtensionManifestPage'
import { ExtensionsAreaPageProps } from '../ExtensionsArea'
import { RegistryExtensionManagePage } from '../registry/RegistryExtensionManagePage'
import { RegistryExtensionNewReleasePage } from '../registry/RegistryExtensionNewReleasePage'
import { RegistryExtensionOverviewPage } from '../registry/RegistryExtensionOverviewPage'
import { ExtensionAreaHeader } from './ExtensionAreaHeader'

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

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

interface Props extends ExtensionsAreaPageProps, RouteComponentProps<{ extensionID: string }> {
    isLightTheme: boolean
}

interface State {
    /** The registry extension, undefined while loading, or an error.  */
    extensionOrError?: ConfiguredExtension<GQL.IRegistryExtension> | ErrorLike
}

/**
 * Properties passed to all page components in the registry extension area.
 */
export interface ExtensionAreaPageProps extends ConfigurationCascadeProps, ExtensionsProps {
    /** The extension registry area main URL. */
    url: string

    /** The extension that is the subject of the page. */
    extension: ConfiguredExtension<GQL.IRegistryExtension>

    /** Called when the component updates the extension and it should be refreshed here. */
    onDidUpdateExtension: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * An extension area.
 */
export class ExtensionArea extends React.Component<Props> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
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
                        type PartialStateUpdate = Pick<State, 'extensionOrError'>
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

    public componentWillReceiveProps(props: Props): void {
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
                <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.extensionOrError.message)} />
            )
        }

        const url = this.props.match.url.replace(/\/-\/?$/, '')

        const transferProps: ExtensionAreaPageProps = {
            url,
            authenticatedUser: this.props.authenticatedUser,
            onDidUpdateExtension: this.onDidUpdateExtension,
            configurationCascade: this.props.configurationCascade,
            extension: this.state.extensionOrError,
            extensions: this.props.extensions,
        }

        return (
            <div className="registry-extension-area area--vertical">
                <ExtensionAreaHeader {...this.props} {...transferProps} />
                <div className="container pt-3">
                    <Switch>
                        <Route
                            path={url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RegistryExtensionOverviewPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${url}/-/manifest`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RegistryExtensionManifestPage
                                    {...routeComponentProps}
                                    {...transferProps}
                                    isLightTheme={this.props.isLightTheme}
                                />
                            )}
                        />
                        <Route
                            path={`${url}/-/manage`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RegistryExtensionManagePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${url}/-/releases/new`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RegistryExtensionNewReleasePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }

    private onDidUpdateExtension = () => this.refreshRequests.next()
}
