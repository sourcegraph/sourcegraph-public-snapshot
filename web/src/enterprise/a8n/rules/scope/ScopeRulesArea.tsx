import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { HeroPage } from '../../../../components/HeroPage'
import { AutomationIcon } from '../../icons'
import { RuleScope } from '../types'
import { RulesListPage } from './list/RulesListPage'

export interface RulesAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The rule scope. */
    scope: RuleScope

    /** The URL to the rules area. */
    rulesURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface Props extends Pick<RulesAreaContext, Exclude<keyof RulesAreaContext, 'rulesURL'>>, RouteComponentProps<{}> {}

/**
 * The rules area for a particular scope.
 */
export const RulesArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: RulesAreaContext = {
        ...props,
        rulesURL: match.url,
    }
    return (
        <div className="container">
            <Switch>
                <Route path={match.url} exact={true}>
                    <h1 className="h2 my-3 d-flex align-items-center font-weight-normal">
                        <AutomationIcon className="icon-inline mr-3" /> Rules
                    </h1>
                    <RulesListPage {...context} />
                </Route>
                <Route
                    path={`${match.url}/:id`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ id: string }>) => (
                        <RuleArea {...context} ruleID={routeComponentProps.match.params.id} />
                    )}
                />
                <Route>
                    <HeroPage
                        icon={MapSearchIcon}
                        title="404: Not Found"
                        subtitle="Sorry, the requested page was not found."
                    />
                </Route>
            </Switch>
        </div>
    )
}
