import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { NamespaceProps } from '../../namespaces'
import { NamespaceArea } from '../../namespaces/NamespaceArea'
import { ThemeProps } from '../../theme'
import { OrgSavedSearchesCreateForm } from '../saved-searches/OrgSavedSearchesCreateForm'
import { OrgSavedSearchesUpdateForm } from '../saved-searches/OrgSavedSearchesUpdateForm'
import { OrgSavedSearchListPage } from '../saved-searches/OrgSavedSearchListPage'
import { OrgSettingsArea } from '../settings/OrgSettingsArea'
import { OrgHeader } from './OrgHeader'
import { OrgInvitationPage } from './OrgInvitationPage'
import { OrgMembersPage } from './OrgMembersPage'
import { OrgOverviewPage } from './OrgOverviewPage'

function queryOrganization(args: { name: string }): Observable<GQL.IOrg | null> {
    return queryGraphQL(
        gql`
            query Organization($name: String!) {
                organization(name: $name) {
                    __typename
                    id
                    name
                    displayName
                    url
                    settingsURL
                    viewerPendingInvitation {
                        id
                        sender {
                            username
                            displayName
                            avatarURL
                            createdAt
                        }
                        respondURL
                    }
                    viewerIsMember
                    viewerCanAdminister
                    createdAt
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.organization) {
                throw createAggregateError(errors)
            }
            return data.organization
        })
    )
}

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested organization was not found." />
)

interface Props
    extends RouteComponentProps<{ name: string }>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        ExtensionsControllerProps {
    /**
     * The currently authenticated user.
     */
    authenticatedUser: GQL.IUser | null
}

interface State {
    /**
     * The fetched org or an error if an error occurred; undefined while loading.
     */
    orgOrError?: GQL.IOrg | ErrorLike
}

/**
 * Properties passed to all page components in the org area.
 */
export interface OrgAreaPageProps
    extends PlatformContextProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        NamespaceProps {
    /** The org that is the subject of the page. */
    org: GQL.IOrg

    /** Called when the organization is updated and must be reloaded. */
    onOrganizationUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * An organization's public profile area.
 */
export class OrgArea extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<{ name: string }>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route-matched org name.
        const nameChanges = this.routeMatchChanges.pipe(
            map(({ name }) => name),
            distinctUntilChanged()
        )

        // Fetch organization.
        this.subscriptions.add(
            combineLatest(nameChanges, merge(this.refreshRequests.pipe(mapTo(false)), of(true)))
                .pipe(
                    switchMap(([name, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'orgOrError'>
                        return queryOrganization({ name }).pipe(
                            catchError(error => [error]),
                            map((c): PartialStateUpdate => ({ orgOrError: c })),

                            // Don't clear old org data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? { orgOrError: undefined } : {})
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
        if (!this.state.orgOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.orgOrError)) {
            return (
                <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(this.state.orgOrError.message)} />
            )
        }

        const transferProps: OrgAreaPageProps = {
            authenticatedUser: this.props.authenticatedUser,
            org: this.state.orgOrError,
            onOrganizationUpdate: this.onDidUpdateOrganization,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
            extensionsController: this.props.extensionsController,
            namespace: this.state.orgOrError,
        }

        if (this.props.location.pathname === `${this.props.match.url}/invitation`) {
            // The OrgInvitationPage is displayed without the OrgHeader because it is modal-like.
            return <OrgInvitationPage {...transferProps} onDidRespondToInvitation={this.onDidRespondToInvitation} />
        }

        return (
            <div className="org-area w-100">
                <OrgHeader {...this.props} {...transferProps} className="border-bottom mt-4" />
                <div className="container mt-3">
                    <ErrorBoundary location={this.props.location}>
                        <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                            <Switch>
                                <Route
                                    path={this.props.match.url}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgOverviewPage {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/members`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgMembersPage {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/searches`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgSavedSearchListPage {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/searches/add`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgSavedSearchesCreateForm {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/searches/:id`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgSavedSearchesUpdateForm {...routeComponentProps} {...transferProps} />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/settings`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <OrgSettingsArea
                                            {...routeComponentProps}
                                            {...transferProps}
                                            isLightTheme={this.props.isLightTheme}
                                        />
                                    )}
                                />
                                <Route
                                    path={`${this.props.match.url}/namespace`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <NamespaceArea
                                            {...routeComponentProps}
                                            {...transferProps}
                                            isLightTheme={this.props.isLightTheme}
                                        />
                                    )}
                                />

                                {/* Redirect from previous /users/:username/account -> /users/:username/settings/profile. */}
                                <Route
                                    path={`${this.props.match.url}/account`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={() => <Redirect to={`${this.props.match.url}/settings/profile`} />}
                                />

                                <Route key="hardcoded-key" component={NotFoundPage} />
                            </Switch>
                        </React.Suspense>
                    </ErrorBoundary>
                </div>
            </div>
        )
    }

    private onDidRespondToInvitation = () => this.refreshRequests.next()

    private onDidUpdateOrganization = () => this.refreshRequests.next()
}
