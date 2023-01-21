import * as React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorLike, isErrorLike, asError, logger } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, ErrorMessage } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { requestGraphQL } from '../../backend/graphql'
import { BatchChangesProps } from '../../batches'
import { BreadcrumbsProps, BreadcrumbSetters } from '../../components/Breadcrumbs'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { Page } from '../../components/Page'
import { OrganizationResult, OrganizationVariables, OrgAreaOrganizationFields } from '../../graphql-operations'
import { NamespaceProps } from '../../namespaces'
import { RouteDescriptor } from '../../util/contributions'
import { OrgSettingsAreaRoute } from '../settings/OrgSettingsArea'
import { OrgSettingsSidebarItems } from '../settings/OrgSettingsSidebar'

import { OrgAreaHeaderNavItem, OrgHeader } from './OrgHeader'
import { OrgInvitationPageLegacy } from './OrgInvitationPageLegacy'

function queryOrganization(args: {
    name: string
    // id: string
    // flagName: string organizationFeatureFlagValue(orgID: $orgID, flagName: $flagName)
}): Observable<OrgAreaOrganizationFields> {
    return requestGraphQL<OrganizationResult, OrganizationVariables>(
        gql`
            query Organization($name: String!) {
                organization(name: $name) {
                    ...OrgAreaOrganizationFields
                }
            }

            fragment OrgAreaOrganizationFields on Org {
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
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.organization) {
                throw new Error(`Organization not found: ${JSON.stringify(args.name)}`)
            }
            return data.organization
        })
    )
}

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested organization was not found." />
)

export interface OrgAreaRoute extends RouteDescriptor<OrgAreaRouteContext> {
    /** When true, the header is not rendered and the component is not wrapped in a container. */
    fullPage?: boolean
}

export interface OrgAreaProps
    extends RouteComponentProps<{ name: string }>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    orgAreaRoutes: readonly OrgAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]

    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

interface State extends BreadcrumbSetters {
    /**
     * The fetched org or an error if an error occurred; undefined while loading.
     */
    orgOrError?: OrgAreaOrganizationFields | ErrorLike
}

/**
 * Properties passed to all page components in the org area.
 */
export interface OrgAreaRouteContext
    extends PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        NamespaceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    /** The org that is the subject of the page. */
    org: OrgAreaOrganizationFields

    /** Called when the organization is updated and must be reloaded. */
    onOrganizationUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser

    isSourcegraphDotCom: boolean

    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]
}

/**
 * An organization's public profile area.
 */
export class OrgArea extends React.Component<OrgAreaProps> {
    public state: State

    private componentUpdates = new Subject<OrgAreaProps>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: OrgAreaProps) {
        super(props)
        this.state = {
            setBreadcrumb: props.setBreadcrumb,
            useBreadcrumb: props.useBreadcrumb,
        }
    }

    public componentDidMount(): void {
        // Changes to the route-matched org name.
        const nameChanges = this.componentUpdates.pipe(
            map(props => props.match.params.name),
            distinctUntilChanged()
        )

        // Fetch organization.
        this.subscriptions.add(
            combineLatest([nameChanges, merge(this.refreshRequests.pipe(mapTo(false)), of(true))])
                .pipe(
                    switchMap(([name, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'orgOrError'>
                        return queryOrganization({ name }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            map((orgOrError): PartialStateUpdate => ({ orgOrError })),
                            // Don't clear old org data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? { orgOrError: undefined } : {})
                        )
                    })
                )
                .subscribe(
                    stateUpdate => {
                        if (stateUpdate.orgOrError && !isErrorLike(stateUpdate.orgOrError)) {
                            const childBreadcrumbSetters = this.props.setBreadcrumb({
                                key: 'OrgArea',
                                link: { to: stateUpdate.orgOrError.url, label: stateUpdate.orgOrError.name },
                            })
                            this.subscriptions.add(childBreadcrumbSetters)
                            this.setState({
                                useBreadcrumb: childBreadcrumbSetters.useBreadcrumb,
                                setBreadcrumb: childBreadcrumbSetters.setBreadcrumb,
                                orgOrError: stateUpdate.orgOrError,
                            })
                        } else {
                            this.setState(stateUpdate)
                        }
                    },
                    error => logger.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
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
                <HeroPage
                    icon={AlertCircleIcon}
                    title="Error"
                    subtitle={<ErrorMessage error={this.state.orgOrError} />}
                />
            )
        }

        const context: OrgAreaRouteContext = {
            authenticatedUser: this.props.authenticatedUser,
            org: this.state.orgOrError,
            onOrganizationUpdate: this.onDidUpdateOrganization,
            platformContext: this.props.platformContext,
            settingsCascade: this.props.settingsCascade,
            isLightTheme: this.props.isLightTheme,
            namespace: this.state.orgOrError,
            telemetryService: this.props.telemetryService,
            isSourcegraphDotCom: this.props.isSourcegraphDotCom,
            batchChangesEnabled: this.props.batchChangesEnabled,
            batchChangesExecutionEnabled: this.props.batchChangesExecutionEnabled,
            batchChangesWebhookLogsEnabled: this.props.batchChangesWebhookLogsEnabled,
            breadcrumbs: this.props.breadcrumbs,
            setBreadcrumb: this.state.setBreadcrumb,
            useBreadcrumb: this.state.useBreadcrumb,
            orgSettingsAreaRoutes: this.props.orgSettingsAreaRoutes,
            orgSettingsSideBarItems: this.props.orgSettingsSideBarItems,
        }

        if (this.props.location.pathname === `${this.props.match.url}/invitation`) {
            // The OrgInvitationPageLegacy is displayed without the OrgHeader because it is modal-like.
            return <OrgInvitationPageLegacy {...context} onDidRespondToInvitation={this.onDidRespondToInvitation} />
        }

        return (
            <ErrorBoundary location={this.props.location}>
                <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                    <Switch>
                        {this.props.orgAreaRoutes.map(
                            ({ path, exact, render, condition = () => true, fullPage }) =>
                                condition(context) && (
                                    <Route
                                        path={this.props.match.url + path}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        exact={exact}
                                        render={routeComponentProps =>
                                            fullPage ? (
                                                render({ ...context, ...routeComponentProps })
                                            ) : (
                                                <Page className="org-area">
                                                    <OrgHeader
                                                        {...this.props}
                                                        {...context}
                                                        navItems={this.props.orgAreaHeaderNavItems}
                                                        className="mb-3"
                                                    />
                                                    <div className="container">
                                                        {render({ ...context, ...routeComponentProps })}
                                                    </div>
                                                </Page>
                                            )
                                        }
                                    />
                                )
                        )}
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </React.Suspense>
            </ErrorBoundary>
        )
    }

    private onDidRespondToInvitation = (accepted: boolean): void => {
        if (!accepted) {
            this.props.history.push('/user/settings')
            return
        }
        this.refreshRequests.next()
    }

    private onDidUpdateOrganization = (): void => this.refreshRequests.next()
}
