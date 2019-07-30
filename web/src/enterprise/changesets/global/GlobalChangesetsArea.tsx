import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { RepositoryChangesetsAreaContext } from '../repository/RepositoryChangesetsArea'
import { GlobalChangesetsListPage } from './list/GlobalChangesetsListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            RepositoryChangesetsAreaContext,
            Exclude<keyof RepositoryChangesetsAreaContext, 'changesetsURL' | 'repository' | 'setBreadcrumbItem'>
        > {}

/**
 * The global changesets area.
 */
export const GlobalChangesetsArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: Pick<RepositoryChangesetsAreaContext, Exclude<keyof RepositoryChangesetsAreaContext, 'repository'>> = {
        ...props,
        changesetsURL: match.url,
    }
    return (
        <Switch>
            <Route path={context.changesetsURL} exact={true}>
                <div className="container mt-4">
                    <GlobalChangesetsListPage {...context} />
                </div>
            </Route>
        </Switch>
    )
}
