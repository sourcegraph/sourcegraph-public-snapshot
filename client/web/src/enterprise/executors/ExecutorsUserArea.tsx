import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { HeroPage } from '../../components/HeroPage'
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
            <CompatRoute
                path={`${match.url}/secrets`}
                render={(props: RouteComponentProps) => (
                    <UserExecutorSecretsListPage userID={outerProps.namespaceID} {...outerProps} {...props} />
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
