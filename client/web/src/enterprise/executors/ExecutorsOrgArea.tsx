import React from 'react'

import { Route, RouteComponentProps, Switch } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NotFoundPage } from '../../components/HeroPage'
import { Scalars } from '../../graphql-operations'

import type { OrgExecutorSecretsListPageProps } from './secrets/ExecutorSecretsListPage'

const OrgExecutorSecretsListPage = lazyComponent<OrgExecutorSecretsListPageProps, 'OrgExecutorSecretsListPage'>(
    () => import('./secrets/ExecutorSecretsListPage'),
    'OrgExecutorSecretsListPage'
)

export interface ExecutorsOrgAreaProps<RouteProps extends {} = {}> extends RouteComponentProps<RouteProps> {
    namespaceID: Scalars['ID']
}

/** The page area for all executors settings in org settings. */
export const ExecutorsOrgArea: React.FunctionComponent<React.PropsWithChildren<ExecutorsOrgAreaProps>> = ({
    match,
    ...outerProps
}) => (
    <>
        <Switch>
            <Route
                path={`${match.url}/secrets`}
                render={props => (
                    <OrgExecutorSecretsListPage orgID={outerProps.namespaceID} {...outerProps} {...props} />
                )}
                exact={true}
            />
            <Route render={() => <NotFoundPage pageType="executors" />} key="hardcoded-key" />
        </Switch>
    </>
)
