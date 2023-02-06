import React from 'react'

import { Route, RouteComponentProps, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NotFoundPage } from '../../components/HeroPage'

import type { ExecutorsListPageProps } from './instances/ExecutorsListPage'
import type { GlobalExecutorSecretsListPageProps } from './secrets/ExecutorSecretsListPage'

const ExecutorsListPage = lazyComponent<ExecutorsListPageProps, 'ExecutorsListPage'>(
    () => import('./instances/ExecutorsListPage'),
    'ExecutorsListPage'
)

const GlobalExecutorSecretsListPage = lazyComponent<
    GlobalExecutorSecretsListPageProps,
    'GlobalExecutorSecretsListPage'
>(() => import('./secrets/ExecutorSecretsListPage'), 'GlobalExecutorSecretsListPage')

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
                render={props => <GlobalExecutorSecretsListPage {...outerProps} {...props} />}
                exact={true}
            />
            <Route render={() => <NotFoundPage pageType="settings" />} key="hardcoded-key" />
        </Switch>
    </>
)
