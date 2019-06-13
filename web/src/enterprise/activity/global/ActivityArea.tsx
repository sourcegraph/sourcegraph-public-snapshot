import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { HeroPage } from '../../../components/HeroPage'
import { ActivityTimelinePage } from '../timeline/ActivityTimelinePage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the activity area.
 */
export interface ActivityAreaContext extends ExtensionsControllerProps {
    isLightTheme: boolean
}

export interface EventsAreaProps extends ActivityAreaContext, RouteComponentProps<{}> {}

/**
 * The global activity area.
 */
export const ActivityArea: React.FunctionComponent<EventsAreaProps> = ({ match, ...props }) => {
    const context: ActivityAreaContext = {
        ...props,
    }

    return (
        <div className="container mt-3">
            <Switch>
                <Route
                    path={match.url}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={routeComponentProps => <ActivityTimelinePage {...routeComponentProps} {...context} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
