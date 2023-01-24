import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

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

export interface ExecutorsSiteAdminAreaProps<RouteProps extends {} = {}> extends RouteComponentProps<RouteProps> {}

/** The page area for all executors settings in site-admin. */
export const ExecutorsSiteAdminArea: React.FunctionComponent<React.PropsWithChildren<ExecutorsSiteAdminAreaProps>> = ({
    match,
    ...outerProps
}) => (
    <>
        <Switch>
            <CompatRoute
                render={(props: RouteComponentProps<{}>) => <ExecutorsListPage {...outerProps} {...props} />}
                path={match.url}
                exact={true}
            />
            <CompatRoute
                path={`${match.url}/secrets`}
                render={(props: RouteComponentProps<{}>) => (
                    <GlobalExecutorSecretsListPage {...outerProps} {...props} />
                )}
                exact={true}
            />
            <CompatRoute component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    </>
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)
