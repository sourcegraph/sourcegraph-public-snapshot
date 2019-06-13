import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { HeroPage } from '../components/HeroPage'
import { NamespaceProjectsPage } from './projects/NamespaceProjectsPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends ExtensionsControllerProps {
    namespace: Pick<GQL.Namespace, '__typename' | 'id'>
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface NamespaceAreaProps extends NamespaceAreaContext, RouteComponentProps<{}> {}

/**
 * The namespace area.
 */
export const NamespaceArea: React.FunctionComponent<NamespaceAreaProps> = ({ match, ...props }) => {
    const context: NamespaceAreaContext = props
    return (
        <div className="container mt-3">
            <Switch>
                <Route
                    path={`${match.url}/projects`}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={routeComponentProps => <NamespaceProjectsPage {...routeComponentProps} {...context} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
