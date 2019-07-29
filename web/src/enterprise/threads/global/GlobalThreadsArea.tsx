import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ThreadArea } from '../detail/ThreadArea'
import { NamespaceThreadsAreaContext } from '..../repository/NamespaceThreadsArea
import { GlobalThreadsListPage } from './list/GlobalThreadsListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            NamespaceThreadsAreaContext,
            Exclude<keyof NamespaceThreadsAreaContext, 'threadsURL' | 'namespace' | 'setBreadcrumbItem'>
        > {}

/**
 * The global threads area.
 */
export const GlobalThreadsArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: Pick<NamespaceThreadsAreaContext, Exclude<keyof NamespaceThreadsAreaContext, 'namespace'>> = {
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
