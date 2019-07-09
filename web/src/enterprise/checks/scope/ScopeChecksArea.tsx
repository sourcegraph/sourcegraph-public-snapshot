import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import * as sourcegraph from 'sourcegraph'
import { CheckID } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { ChecklistIcon } from '../../../util/octicons'
import { CheckArea } from '../detail/CheckArea'
import { ChecksListPage } from './list/ChecksListPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

export interface ChecksAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The check scope. */
    scope: sourcegraph.CheckScope | sourcegraph.WorkspaceRoot

    /** The URL to the checks area. */
    checksURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface Props extends ChecksAreaContext, RouteComponentProps<{}> {}

/**
 * The checks area for a particular scope.
 */
export const ChecksArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: ChecksAreaContext = {
        ...props,
        checksURL: match.url,
    }
    return (
        <Switch>
            <Route path={match.url} exact={true}>
                <div className="container">
                    <h1 className="h2 my-3 d-flex align-items-center font-weight-normal">
                        <ChecklistIcon className="icon-inline mr-3" /> Checks
                    </h1>
                    <ChecksListPage {...context} />
                </div>
            </Route>
            <Route
                path={`${match.url}/:type/:id`}
                // tslint:disable-next-line:jsx-no-lambda
                render={(routeComponentProps: RouteComponentProps<CheckID>) => (
                    <CheckArea
                        {...context}
                        checkID={routeComponentProps.match.params}
                        checkURL={routeComponentProps.match.url}
                    />
                )}
            />
            <Route key="hardcoded-key" component={NotFoundPage} />
        </Switch>
    )
}
