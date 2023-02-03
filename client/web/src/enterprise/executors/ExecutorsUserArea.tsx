import React from 'react'

import { Route, RouteComponentProps, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NotFoundPage } from '../../components/HeroPage'
import { Scalars } from '../../graphql-operations'

import type { UserExecutorSecretsListPageProps } from './secrets/ExecutorSecretsListPage'

const UserExecutorSecretsListPage = lazyComponent<UserExecutorSecretsListPageProps, 'UserExecutorSecretsListPage'>(
    () => import('./secrets/ExecutorSecretsListPage'),
    'UserExecutorSecretsListPage'
)

export interface ExecutorsUserAreaProps<RouteProps extends {} = {}> extends RouteComponentProps<RouteProps> {
    namespaceID: Scalars['ID']
}

/** The page area for all executors settings in user settings. */
export const ExecutorsUserArea: React.FunctionComponent<React.PropsWithChildren<ExecutorsUserAreaProps>> = ({
    match,
    ...outerProps
}) => (
    <>
        <Switch>
            <Route
                path={`${match.url}/secrets`}
                render={props => (
                    <UserExecutorSecretsListPage userID={outerProps.namespaceID} {...outerProps} {...props} />
                )}
                exact={true}
            />
            <Route render={() => <NotFoundPage pageType="settings" />} key="hardcoded-key" />
        </Switch>
    </>
)
