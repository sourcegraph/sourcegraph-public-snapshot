import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'
import { RouteDescriptor } from '../util/contributions'
import { SiteAdminSidebar, SiteAdminSideBarGroups } from './SiteAdminSidebar'

const NotFoundPage: React.ComponentType<{}> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested site admin page was not found."
    />
)

const NotSiteAdminPage: React.ComponentType<{}> = () => (
    <HeroPage icon={MapSearchIcon} title="403: Forbidden" subtitle="Only site admins are allowed here." />
)

export interface SiteAdminAreaRouteContext extends ExtensionsProps {
    site: Pick<GQL.ISite, '__typename' | 'id'>
    authenticatedUser: GQL.IUser
    isLightTheme: boolean

    /** This property is only used by {@link SiteAdminOverviewPage}. */
    overviewComponents: ReadonlyArray<React.ComponentType>
}

export interface SiteAdminAreaRoute extends RouteDescriptor<SiteAdminAreaRouteContext> {}

interface SiteAdminAreaProps extends RouteComponentProps<{}>, ExtensionsProps {
    routes: ReadonlyArray<SiteAdminAreaRoute>
    sideBarGroups: SiteAdminSideBarGroups
    overviewComponents: ReadonlyArray<React.ComponentType>
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display site admin information.
 */
export const SiteAdminArea = withAuthenticatedUser(
    class SiteAdminArea extends React.Component<SiteAdminAreaProps> {
        public render(): JSX.Element | null {
            // If not site admin, redirect to sign in.
            if (!this.props.authenticatedUser.siteAdmin) {
                return <NotSiteAdminPage />
            }

            const context: SiteAdminAreaRouteContext = {
                authenticatedUser: this.props.authenticatedUser,
                extensions: this.props.extensions,
                isLightTheme: this.props.isLightTheme,
                site: { __typename: 'Site' as 'Site', id: window.context.siteGQLID },
                overviewComponents: this.props.overviewComponents,
            }

            return (
                <div className="site-admin-area area">
                    <SiteAdminSidebar className="area__sidebar" groups={this.props.sideBarGroups} />
                    <div className="area__content">
                        <Switch>
                            {this.props.routes.map(
                                ({ render, path, exact, condition = () => true }) =>
                                    condition(context) && (
                                        <Route
                                            // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            key="hardcoded-key"
                                            path={this.props.match.url + path}
                                            exact={exact}
                                            // tslint:disable-next-line:jsx-no-lambda RouteProps.render is an exception
                                            render={routeComponentProps =>
                                                render({ ...context, ...routeComponentProps })
                                            }
                                        />
                                    )
                            )}
                            <Route component={NotFoundPage} />
                        </Switch>
                    </div>
                </div>
            )
        }
    }
)
