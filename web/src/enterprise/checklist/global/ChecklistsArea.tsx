import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { ChecklistArea } from '../detail/ChecklistArea'
import { ChecklistsListPage } from '../list/ChecklistsListPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle={`Sorry, the requested page was not found.`} />
)

/**
 * Properties passed to all page components in the checklist area.
 */
export interface ChecklistsAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    isLightTheme: boolean
}

export interface ChecklistsAreaProps extends ChecklistsAreaContext, RouteComponentProps<{}> {}

/**
 * The global checklist area.
 */
export const ChecklistsArea: React.FunctionComponent<ChecklistsAreaProps> = ({ match, ...props }) => {
    const context: ChecklistsAreaContext = {
        ...props,
    }

    return (
        <Switch>
            <Route
                path={match.url}
                exact={true}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <ChecklistsListPage {...routeComponentProps} {...context} />}
            />
            <Route
                path={`${match.url}/:checklistID`}
                // tslint:disable-next-line:jsx-no-lambda
                render={routeComponentProps => <ChecklistArea {...routeComponentProps} {...context} />}
            />
            <Route key="hardcoded-key" component={NotFoundPage} />
        </Switch>
    )
}
