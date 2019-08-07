import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { RepositoryThreadsAreaContext } from '../repository/RepositoryThreadsArea'
import { GlobalThreadsListPage } from './list/GlobalThreadsListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            RepositoryThreadsAreaContext,
            Exclude<keyof RepositoryThreadsAreaContext, 'threadsURL' | 'repo' | 'setBreadcrumbItem'>
        > {}

/**
 * The global threads area.
 */
export const GlobalThreadsArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: Pick<RepositoryThreadsAreaContext, Exclude<keyof RepositoryThreadsAreaContext, 'repo'>> = {
        ...props,
        threadsURL: match.url,
    }
    return (
        <Switch>
            <Route path={context.threadsURL} exact={true}>
                <div className="container mt-4">
                    <GlobalThreadsListPage {...context} />
                </div>
            </Route>
        </Switch>
    )
}
