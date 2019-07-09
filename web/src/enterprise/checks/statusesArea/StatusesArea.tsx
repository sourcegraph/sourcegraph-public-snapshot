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
import { CheckArea } from '../../checks/statusArea/CheckArea'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

export interface ChecksAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The status scope. */
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot

    /** The URL to the statuses area. */
    statusesURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface Props extends ChecksAreaContext, RouteComponentProps<{}> {}

/**
 * The statuses area.
 */
export const ChecksArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: ChecksAreaContext = {
        ...props,
        statusesURL: match.url,
    }
    return (
        <Switch>
            <Route path={match.url} exact={true}>
                <div className="container">
                    <h1 className="h2 my-3 d-flex align-items-center font-weight-normal">
                        <ChecklistIcon className="icon-inline mr-3" /> Status
                    </h1>
                    <CombinedStatusPage {...context} />
                </div>
            </Route>
            <Route
                path={`${match.url}/:name`}
                // tslint:disable-next-line:jsx-no-lambda
                render={(routeComponentProps: RouteComponentProps<{ name: string }>) => (
                    <CheckArea
                        {...context}
                        name={routeComponentProps.match.params.name}
                        statusURL={routeComponentProps.match.url}
                    />
                )}
            />
            <Route key="hardcoded-key" component={NotFoundPage} />
        </Switch>
    )
}
