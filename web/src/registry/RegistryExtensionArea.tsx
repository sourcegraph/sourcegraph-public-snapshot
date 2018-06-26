import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { RegistryAreaPageProps } from './RegistryArea'
import { RegistryExtensionContributionsPage } from './RegistryExtensionContributionsPage'
import { RegistryExtensionEditPage } from './RegistryExtensionEditPage'
import { RegistryExtensionHeader } from './RegistryExtensionHeader'
import { RegistryExtensionManifestPage } from './RegistryExtensionManifestPage'
import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'
import { registryExtensionFragment } from './RegistryExtensionsPage'
import { RegistryExtensionUsersPage } from './RegistryExtensionUsersPage'

function queryRegistryExtension(args: { extensionID: string }): Observable<GQL.IRegistryExtension | null> {
    return queryGraphQL(
        gql`
            query RegistryExtension($extensionID: String!) {
                extensionRegistry {
                    extension(extensionID: $extensionID) {
                        ...RegistryExtensionFields
                    }
                }
            }
            ${registryExtensionFragment}
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.extensionRegistry || !data.extensionRegistry.extension) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extension
        })
    )
}

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

interface Props extends RegistryAreaPageProps, RouteComponentProps<{ extensionID: string }> {
    isLightTheme: boolean
}

interface State {
    /** The registry extension, undefined while loading, or an error.  */
    extensionOrError?: GQL.IRegistryExtension | ErrorLike
}

/**
 * Properties passed to all page components in the registry extension area.
 */
export interface RegistryExtensionAreaPageProps {
    /** The extension registry area main URL. */
    url: string

    /** The extension that is the subject of the page. */
    extension: GQL.IRegistryExtension

    /** Called when the component updates the extension and it should be refreshed here. */
    onDidUpdateExtension: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * A registry extension area.
 */
export class RegistryExtensionArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ extensionID: string }>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route-matched extension ID.
        const extensionIDChanges = this.routeMatchChanges.pipe(
            map(({ extensionID }) => extensionID),
            distinctUntilChanged()
        )

        // Fetch extension.
        this.subscriptions.add(
            combineLatest(extensionIDChanges, merge(this.refreshRequests.pipe(mapTo(true)), of(false)))
                .pipe(
                    switchMap(([extensionID, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'extensionOrError'>
                        return queryRegistryExtension({ extensionID }).pipe(
                            catchError(error => [error]),
                            map(c => ({ extensionOrError: c } as PartialStateUpdate)),

                            // Don't clear old data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? {} : { extensionOrError: undefined })
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.routeMatchChanges.next(this.props.match.params)
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
        if (!this.state.extensionOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.extensionOrError)) {
            return (
                <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.extensionOrError.message)} />
            )
        }

        const url = this.props.match.url.replace(/\/-\/?$/, '')

        const transferProps: RegistryExtensionAreaPageProps = {
            url,
            authenticatedUser: this.props.authenticatedUser,
            onDidUpdateExtension: this.onDidUpdateExtension,
            extension: this.state.extensionOrError,
        }

        return (
            <div className="registry-extension-area area--vertical">
                <RegistryExtensionHeader className="area--vertical__header" {...this.props} {...transferProps} />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
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
                                path={`${url}/-/contributions`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RegistryExtensionContributionsPage {...routeComponentProps} {...transferProps} />
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
                                path={`${url}/-/users`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RegistryExtensionUsersPage {...routeComponentProps} {...transferProps} />
                                )}
                            />
                            <Route
                                path={`${url}/-/edit`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RegistryExtensionEditPage
                                        {...routeComponentProps}
                                        {...transferProps}
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                )}
                            />
                            <Route key="hardcoded-key" component={NotFoundPage} />
                        </Switch>
                    </div>
                </div>
            </div>
        )
    }

    private onDidUpdateExtension = () => this.refreshRequests.next()
}
