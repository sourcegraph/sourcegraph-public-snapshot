import React, { useMemo } from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters, BreadcrumbsProps, Breadcrumbs } from '../../../components/Breadcrumbs'
import { Scalars } from '../../../../../shared/src/graphql-operations'
import { lazyComponent } from '../../../util/lazyComponent'
import { UserCampaignListPageProps, CampaignListPageProps, OrgCampaignListPageProps } from '../list/CampaignListPage'
import { CampaignApplyPageProps } from '../apply/CampaignApplyPage'
import { CreateCampaignPageProps } from '../create/CreateCampaignPage'
import { CampaignDetailsPageProps } from '../detail/CampaignDetailsPage'
import { CampaignClosePageProps } from '../close/CampaignClosePage'
import { CampaignsDotComPageProps } from './marketing/CampaignsDotComPage'
import { Link } from '../../../../../shared/src/components/Link'

const CampaignListPage = lazyComponent<CampaignListPageProps, 'CampaignListPage'>(
    () => import('../list/CampaignListPage'),
    'CampaignListPage'
)
const OrgCampaignListPage = lazyComponent<OrgCampaignListPageProps, 'OrgCampaignListPage'>(
    () => import('../list/CampaignListPage'),
    'OrgCampaignListPage'
)
const UserCampaignListPage = lazyComponent<UserCampaignListPageProps, 'UserCampaignListPage'>(
    () => import('../list/CampaignListPage'),
    'UserCampaignListPage'
)
const CampaignApplyPage = lazyComponent<CampaignApplyPageProps, 'CampaignApplyPage'>(
    () => import('../apply/CampaignApplyPage'),
    'CampaignApplyPage'
)
const CreateCampaignPage = lazyComponent<CreateCampaignPageProps, 'CreateCampaignPage'>(
    () => import('../create/CreateCampaignPage'),
    'CreateCampaignPage'
)
const CampaignDetailsPage = lazyComponent<CampaignDetailsPageProps, 'CampaignDetailsPage'>(
    () => import('../detail/CampaignDetailsPage'),
    'CampaignDetailsPage'
)
const CampaignClosePage = lazyComponent<CampaignClosePageProps, 'CampaignClosePage'>(
    () => import('../close/CampaignClosePage'),
    'CampaignClosePage'
)
const CampaignsDotComPage = lazyComponent<CampaignsDotComPageProps, 'CampaignsDotComPage'>(
    () => import('./marketing/CampaignsDotComPage'),
    'CampaignsDotComPage'
)

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps,
        BreadcrumbsProps,
        BreadcrumbSetters {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea: React.FunctionComponent<Props> = props => {
    if (props.isSourcegraphDotCom) {
        return <CampaignsDotComPage />
    }
    return <AuthenticatedCampaignsArea {...props} />
}

interface AuthenticatedProps extends Props {
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedCampaignsArea = withAuthenticatedUser<AuthenticatedProps>(({ match, ...outerProps }) => {
    const breadcrumbSetters = outerProps.useBreadcrumb(
        useMemo(
            () => ({
                key: 'CampaignsArea',
                element: <Link to={match.url}>Campaigns</Link>,
            }),
            [match.url]
        )
    )
    return (
        <div className="w-100 web-content">
            <Breadcrumbs breadcrumbs={outerProps.breadcrumbs} />
            <div className="container">
                {/* eslint-disable react/jsx-no-bind */}
                <Switch>
                    <Route
                        render={props => <CampaignListPage {...outerProps} {...props} {...breadcrumbSetters} />}
                        path={match.url}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/create`}
                        render={props => <CreateCampaignPage {...outerProps} {...props} {...breadcrumbSetters} />}
                        exact={true}
                    />
                </Switch>
                {/* eslint-enable react/jsx-no-bind */}
            </div>
        </div>
    )
})

export interface UserCampaignsAreaProps extends Props {
    userID: Scalars['ID']
}

export const UserCampaignsArea = withAuthenticatedUser<
    UserCampaignsAreaProps & { authenticatedUser: AuthenticatedUser }
>(({ match, userID, ...outerProps }) => {
    const breadcrumbSetters = outerProps.useBreadcrumb(
        useMemo(
            () => ({
                key: 'CampaignsArea',
                element: <Link to={match.url}>Campaigns</Link>,
            }),
            [match.url]
        )
    )
    if (outerProps.isSourcegraphDotCom || !window.context.campaignsEnabled) {
        return <></>
    }
    return (
        <div className="w-100 web-content">
            <Breadcrumbs breadcrumbs={outerProps.breadcrumbs} />
            <Switch>
                {/* eslint-disable react/jsx-no-bind */}
                <Route
                    path={`${match.url}/apply/:specID`}
                    render={({ match, ...props }: RouteComponentProps<{ specID: string }>) => (
                        <div className="container">
                            <CampaignApplyPage
                                {...outerProps}
                                {...props}
                                {...breadcrumbSetters}
                                specID={match.params.specID}
                            />
                        </div>
                    )}
                />
                <Route
                    path={`${match.url}/create`}
                    render={props => (
                        <div className="container">
                            <CreateCampaignPage {...outerProps} {...props} {...breadcrumbSetters} />
                        </div>
                    )}
                />
                <Route
                    path={`${match.url}/:campaignName/close`}
                    render={({ match, ...props }: RouteComponentProps<{ campaignName: string }>) => (
                        <div className="container">
                            <CampaignClosePage
                                {...outerProps}
                                {...props}
                                {...breadcrumbSetters}
                                namespaceID={userID}
                                campaignName={match.params.campaignName}
                            />
                        </div>
                    )}
                />
                <Route
                    path={`${match.url}/:campaignName`}
                    render={({ match, ...props }: RouteComponentProps<{ campaignName: string }>) => (
                        <CampaignDetailsPage
                            {...outerProps}
                            {...props}
                            {...breadcrumbSetters}
                            namespaceID={userID}
                            campaignName={match.params.campaignName}
                        />
                    )}
                />
                <Route
                    path={match.url}
                    render={props => (
                        <div className="container">
                            <UserCampaignListPage {...outerProps} {...props} {...breadcrumbSetters} userID={userID} />
                        </div>
                    )}
                />
                {/* eslint-enable react/jsx-no-bind */}
            </Switch>
        </div>
    )
})

export interface OrgCampaignsAreaProps extends Props {
    orgID: Scalars['ID']
}

export const OrgCampaignsArea = withAuthenticatedUser<OrgCampaignsAreaProps & { authenticatedUser: AuthenticatedUser }>(
    ({ match, orgID, ...outerProps }) => {
        const breadcrumbSetters = outerProps.useBreadcrumb(
            useMemo(
                () => ({
                    key: 'CampaignsArea',
                    element: <Link to={match.url}>Campaigns</Link>,
                }),
                [match.url]
            )
        )
        if (outerProps.isSourcegraphDotCom || !window.context.campaignsEnabled) {
            return <></>
        }
        return (
            <div className="w-100 web-content">
                <Breadcrumbs breadcrumbs={outerProps.breadcrumbs} />
                <div className="container">
                    <Switch>
                        {/* eslint-disable react/jsx-no-bind */}
                        <Route
                            path={`${match.url}/apply/:specID`}
                            render={({ match, ...props }: RouteComponentProps<{ specID: string }>) => (
                                <CampaignApplyPage
                                    {...props}
                                    {...outerProps}
                                    {...breadcrumbSetters}
                                    specID={match.params.specID}
                                />
                            )}
                        />
                        <Route
                            path={`${match.url}/create`}
                            render={props => <CreateCampaignPage {...props} {...outerProps} {...breadcrumbSetters} />}
                        />
                        <Route
                            path={`${match.url}/:campaignName/close`}
                            render={({ match, ...props }: RouteComponentProps<{ campaignName: string }>) => (
                                <CampaignClosePage
                                    {...props}
                                    {...outerProps}
                                    {...breadcrumbSetters}
                                    namespaceID={orgID}
                                    campaignName={match.params.campaignName}
                                />
                            )}
                        />
                        <Route
                            path={`${match.url}/:campaignName`}
                            render={({ match, ...props }: RouteComponentProps<{ campaignName: string }>) => (
                                <CampaignDetailsPage
                                    {...props}
                                    {...outerProps}
                                    {...breadcrumbSetters}
                                    namespaceID={orgID}
                                    campaignName={match.params.campaignName}
                                />
                            )}
                        />
                        <Route
                            path={match.url}
                            render={props => (
                                <OrgCampaignListPage {...props} {...outerProps} {...breadcrumbSetters} orgID={orgID} />
                            )}
                            exact={true}
                        />
                        {/* eslint-enable react/jsx-no-bind */}
                    </Switch>
                </div>
            </div>
        )
    }
)
