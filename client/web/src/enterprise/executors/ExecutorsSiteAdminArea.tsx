import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { PageHeader } from '@sourcegraph/wildcard'

import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import type { ExecutorsListPageProps } from './instances/ExecutorsListPage'
import { GlobalExecutorSecretsListPage } from './secrets/ExecutorSecretsListPage'

const ExecutorsListPage = lazyComponent<ExecutorsListPageProps, 'ExecutorsListPage'>(
    () => import('./instances/ExecutorsListPage'),
    'ExecutorsListPage'
)

export interface ExecutorsSiteAdminAreaProps<RouteProps extends {} = {}> extends RouteComponentProps<RouteProps> {}

/** The page area for all executors settings in site-admin. */
export const ExecutorsSiteAdminArea: React.FunctionComponent<React.PropsWithChildren<ExecutorsSiteAdminAreaProps>> = ({
    match,
    ...outerProps
}) => (
    <>
        <Switch>
            <Route render={props => <ExecutorsListPage {...outerProps} {...props} />} path={match.url} exact={true} />
            <Route
                path={`${match.url}/secrets`}
                render={props => (
                    <GlobalExecutorSecretsListPage
                        headerLine={
                            <>
                                Configure executor secrets that will be available to everyone on the Sourcegraph
                                instance.
                            </>
                        }
                        {...outerProps}
                        {...props}
                    />
                )}
                exact={true}
            />
            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    </>
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)
