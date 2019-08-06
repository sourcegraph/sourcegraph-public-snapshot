import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { RepositoryIssuesAreaContext } from '../repository/RepositoryIssuesArea'
import { GlobalIssuesListPage } from './list/GlobalIssuesListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            RepositoryIssuesAreaContext,
            Exclude<keyof RepositoryIssuesAreaContext, 'issuesURL' | 'repository' | 'setBreadcrumbItem'>
        > {}

/**
 * The global issues area.
 */
export const GlobalIssuesArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: Pick<RepositoryIssuesAreaContext, Exclude<keyof RepositoryIssuesAreaContext, 'repository'>> = {
        ...props,
        issuesURL: match.url,
    }
    return (
        <Switch>
            <Route path={context.issuesURL} exact={true}>
                <div className="container mt-4">
                    <GlobalIssuesListPage {...context} />
                </div>
            </Route>
        </Switch>
    )
}
