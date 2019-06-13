import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { HeroPage } from '../../../components/HeroPage'
import { ChangesListPage } from '../list/ChangesListPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the changes area.
 */
export interface ChangesAreaContext extends ExtensionsControllerProps {
    isLightTheme: boolean
}

export interface ChangesAreaProps extends ChangesAreaContext, RouteComponentProps<{}> {}

/**
 * The global changes area.
 */
export const ChangesArea: React.FunctionComponent<ChangesAreaProps> = ({ match, ...props }) => {
    const context: ChangesAreaContext = {
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
                    render={routeComponentProps => <ChangesListPage {...routeComponentProps} {...context} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
