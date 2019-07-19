import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerNotificationProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { HeroPage } from '../components/HeroPage'
import { RouteDescriptor } from '../util/contributions'
import { NamespaceCampaignsPage } from './campaigns/NamespaceCampaignsPage'
import { NamespaceProjectsPage } from './projects/NamespaceProjectsPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends ExtensionsControllerNotificationProps {
    namespace: Pick<GQL.Namespace, '__typename' | 'id'>
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}

interface Props extends NamespaceAreaContext, RouteComponentProps<{}> {}

/**
 * The namespace area.
 */
export const NamespaceArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: NamespaceAreaContext = props
    return (
        <div className="container mt-3">
            <Switch>
                <Route
                    path={`${match.url}/campaigns`}
                    exact={true}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={routeComponentProps => <NamespaceCampaignsPage {...routeComponentProps} {...context} />}
                />
                <Route
                    path={`${match.url}/projects`}
                    exact={true}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={routeComponentProps => <NamespaceProjectsPage {...routeComponentProps} {...context} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
