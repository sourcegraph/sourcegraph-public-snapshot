import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { HeroPage } from '../../components/HeroPage'

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

interface Props {}

/** The page area for all executors settings in site-admin. */
export const ExecutorsSiteAdminArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ ...outerProps }) => {
    const url = '/site-admin/executors'
    return (
        <>
            <Switch>
                <Route render={props => <ExecutorsListPage {...outerProps} {...props} />} path={url} exact={true} />
                <Route
                    path={`${url}/secrets`}
                    render={props => <GlobalExecutorSecretsListPage {...outerProps} {...props} />}
                    exact={true}
                />
                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </>
    )
}

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)
