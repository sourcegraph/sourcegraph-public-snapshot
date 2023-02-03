import React from 'react'

import { Route, Switch } from 'react-router'

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

const URL = '/site-admin/executors'

/** The page area for all executors settings in site-admin. */
export const ExecutorsSiteAdminArea: React.FC<{}> = () => (
    <>
        <Switch>
            <Route render={() => <ExecutorsListPage />} path={URL} exact={true} />
            <Route
                path={`${URL}/secrets`}
                render={props => <GlobalExecutorSecretsListPage {...props} />}
                exact={true}
            />
            <Route render={() => <NotFoundPage pageType="settings" />} key="hardcoded-key" />
        </Switch>
    </>
)
