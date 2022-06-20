import * as React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike, asError } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { requestGraphQL } from '../../backend/graphql'
import { BatchChangesProps } from '../../batches'
import { BreadcrumbsProps, BreadcrumbSetters } from '../../components/Breadcrumbs'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { Page } from '../../components/Page'
import {
    OrganizationResult,
    OrganizationVariables,
    OrgAreaOrganizationFields,
    OrgFeatureFlagValueResult,
    OrgFeatureFlagValueVariables,
} from '../../graphql-operations'
import { NamespaceProps } from '../../namespaces'
import { RouteDescriptor } from '../../util/contributions'
import { ORG_CODE_FEATURE_FLAG_EMAIL_INVITE } from '../backend'

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
                viewerNeedsCodeHostUpdate
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

function queryMembersFFlag(args: { orgID: string; flagName: string }): Observable<boolean> {
    return requestGraphQL<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        gql`
            query OrgFeatureFlagValue($orgID: ID!, $flagName: String!) {
                organizationFeatureFlagValue(orgID: $orgID, flagName: $flagName)
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.organizationFeatureFlagValue)
    )
}

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested organization was not found." />
)

export interface OrgAreaRoute extends RouteDescriptor<OrgAreaPageProps> {
    /** When true, the header is not rendered and the component is not wrapped in a container. */
    fullPage?: boolean
}

interface Props
    extends RouteComponentProps<{ name: string }>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        ExtensionsControllerProps,
        BatchChangesProps {
    orgAreaRoutes: readonly OrgAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]

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
    newMembersInviteEnabled: boolean
}

/**
 * Properties passed to all page components in the org area.
 */
export interface OrgAreaPageProps
    extends ExtensionsControllerProps,
        PlatformContextProps,
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

    newMembersInviteEnabled: boolean
}

/**
 * An organization's public profile area.
 */
export class OrgArea extends React.Component<Props> {
    public state: State

    private componentUpdates = new Subject<Props>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            setBreadcrumb: props.setBreadcrumb,
            useBreadcrumb: props.useBreadcrumb,
            newMembersInviteEnabled: false,
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
                .pipe(
                    switchMap(state => {
                        const flagObservable =
                            state.orgOrError && !isErrorLike(state.orgOrError)
                                ? queryMembersFFlag({
                                      orgID: state.orgOrError.id,
                                      flagName: ORG_CODE_FEATURE_FLAG_EMAIL_INVITE,
                                  })
                                : of(false)
                        return flagObservable.pipe(
                            catchError((): [boolean] => [false]), // set flag to false in case of error reading it
                            map(newMembersInviteEnabled =>
                                !state.orgOrError
                                    ? { newMembersInviteEnabled }
                                    : { orgOrError: state.orgOrError, newMembersInviteEnabled }
                            )
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
                                newMembersInviteEnabled: stateUpdate.newMembersInviteEnabled,
                            })
                        } else {
                            this.setState(stateUpdate)
                        }
                    },
                    error => console.error(error)
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

        const context: OrgAreaPageProps = {
            authenticatedUser: this.props.authenticatedUser,
            org: this.state.orgOrError,
            onOrganizationUpdate: this.onDidUpdateOrganization,
            extensionsController: this.props.extensionsController,
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
            newMembersInviteEnabled: this.state.newMembersInviteEnabled,
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
