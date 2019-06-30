import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { ChecklistIcon } from '../../../util/octicons'
import { CombinedStatusPage } from '../combinedStatus/CombinedStatusPage'
import { StatusArea } from '../statusArea/StatusArea'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

export interface StatusesAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The status scope. */
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot

    /** The URL to the statuses area. */
    statusesURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface Props extends StatusesAreaContext, RouteComponentProps<{}> {}

/**
 * The statuses area.
 */
export const StatusesArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: StatusesAreaContext = {
        ...props,
        statusesURL: match.url,
    }
    return (
        <div className="w-100">
            <Switch>
                <Route path={match.url} exact={true}>
                    <div className="container-fluid my-5">
                        <h1 className="h2 mb-0 d-flex align-items-center font-weight-normal">
                            <ChecklistIcon className="icon-inline mr-3" /> Status
                        </h1>
                    </div>
                    <CombinedStatusPage {...context} />
                </Route>
                <Route
                    path={`${match.url}/:name`}
                    exact={true}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ name: string }>) => (
                        <StatusArea
                            {...context}
                            name={routeComponentProps.match.params.name}
                            statusURL={routeComponentProps.match.url}
                        />
                    )}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
