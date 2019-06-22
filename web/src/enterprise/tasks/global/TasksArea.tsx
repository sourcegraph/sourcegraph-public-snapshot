import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { TaskArea } from '../detail/TaskArea'
import { TasksListPage } from '../list/TasksListPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the tasks area.
 */
export interface TasksAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    isLightTheme: boolean
}

export interface TasksAreaProps extends TasksAreaContext, RouteComponentProps<{}> {}

/**
 * The global tasks area.
 */
export const TasksArea: React.FunctionComponent<TasksAreaProps> = ({ match, ...props }) => {
    const context: TasksAreaContext = {
        ...props,
    }

    return (
        <Switch>
            <Route
                path={match.url}
                exact={true}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <TasksListPage {...routeComponentProps} {...context} />}
            />
            <Route
                path={`${match.url}/:taskID`}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <TaskArea {...routeComponentProps} {...context} />}
            />
            <Route key="hardcoded-key" component={NotFoundPage} />
        </Switch>
    )
}
