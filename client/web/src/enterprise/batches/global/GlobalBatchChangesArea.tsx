import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { HeroPage } from '../../../components/HeroPage'
import { lazyComponent } from '../../../util/lazyComponent'
import type { BatchChangeClosePageProps } from '../close/BatchChangeClosePage'
import type { CreateBatchChangePageProps } from '../create/CreateBatchChangePage'
import type { BatchChangeDetailsPageProps } from '../detail/BatchChangeDetailsPage'
import type { BatchSpecExecutionDetailsPageProps } from '../execution/BatchSpecExecutionDetailsPage'
import type { BatchChangeListPageProps, NamespaceBatchChangeListPageProps } from '../list/BatchChangeListPage'
import type { BatchChangePreviewPageProps } from '../preview/BatchChangePreviewPage'

import type { DotcomGettingStartedPageProps } from './DotcomGettingStartedPage'

const BatchChangeListPage = lazyComponent<BatchChangeListPageProps, 'BatchChangeListPage'>(
    () => import('../list/BatchChangeListPage'),
    'BatchChangeListPage'
)
const NamespaceBatchChangeListPage = lazyComponent<NamespaceBatchChangeListPageProps, 'NamespaceBatchChangeListPage'>(
    () => import('../list/BatchChangeListPage'),
    'NamespaceBatchChangeListPage'
)
const BatchChangePreviewPage = lazyComponent<BatchChangePreviewPageProps, 'BatchChangePreviewPage'>(
    () => import('../preview/BatchChangePreviewPage'),
    'BatchChangePreviewPage'
)
const CreateBatchChangePage = lazyComponent<CreateBatchChangePageProps, 'CreateBatchChangePage'>(
    () => import('../create/CreateBatchChangePage'),
    'CreateBatchChangePage'
)
const BatchChangeDetailsPage = lazyComponent<BatchChangeDetailsPageProps, 'BatchChangeDetailsPage'>(
    () => import('../detail/BatchChangeDetailsPage'),
    'BatchChangeDetailsPage'
)
const BatchChangeClosePage = lazyComponent<BatchChangeClosePageProps, 'BatchChangeClosePage'>(
    () => import('../close/BatchChangeClosePage'),
    'BatchChangeClosePage'
)
const BatchSpecExecutionDetailsPage = lazyComponent<
    BatchSpecExecutionDetailsPageProps,
    'BatchSpecExecutionDetailsPage'
>(() => import('../execution/BatchSpecExecutionDetailsPage'), 'BatchSpecExecutionDetailsPage')
const DotcomGettingStartedPage = lazyComponent<DotcomGettingStartedPageProps, 'DotcomGettingStartedPage'>(
    () => import('./DotcomGettingStartedPage'),
    'DotcomGettingStartedPage'
)
interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

/**
 * The global batch changes area.
 */
export const GlobalBatchChangesArea: React.FunctionComponent<Props> = props => {
    if (props.isSourcegraphDotCom) {
        return <DotcomGettingStartedPage />
    }
    return <AuthenticatedBatchChangesArea {...props} />
}

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

interface AuthenticatedProps extends Props {
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedBatchChangesArea = withAuthenticatedUser<AuthenticatedProps>(({ match, ...outerProps }) => (
    <div className="w-100">
        <Switch>
            <Route
                render={props => (
                    <BatchChangeListPage headingElement="h1" canCreate={true} {...outerProps} {...props} />
                )}
                path={match.url}
                exact={true}
            />
            <Route
                path={`${match.url}/create`}
                render={props => <CreateBatchChangePage headingElement="h1" {...outerProps} {...props} />}
                exact={true}
            />
            <Route
                path={`${match.url}/executions/:batchSpecID`}
                render={({ match, ...props }: RouteComponentProps<{ batchSpecID: string }>) => (
                    <BatchSpecExecutionDetailsPage
                        {...outerProps}
                        {...props}
                        match={match}
                        batchSpecID={match.params.batchSpecID}
                    />
                )}
            />
            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    </div>
))

export interface NamespaceBatchChangesAreaProps extends Props {
    namespaceID: Scalars['ID']
}

export const NamespaceBatchChangesArea = withAuthenticatedUser<
    NamespaceBatchChangesAreaProps & { authenticatedUser: AuthenticatedUser }
>(({ match, namespaceID, ...outerProps }) => (
    <div className="pb-3">
        <Switch>
            <Route
                path={`${match.url}/apply/:specID`}
                render={({ match, ...props }: RouteComponentProps<{ specID: string }>) => (
                    <BatchChangePreviewPage {...outerProps} {...props} batchSpecID={match.params.specID} />
                )}
            />
            <Route
                path={`${match.url}/:batchChangeName/close`}
                render={({ match, ...props }: RouteComponentProps<{ batchChangeName: string }>) => (
                    <BatchChangeClosePage
                        {...outerProps}
                        {...props}
                        namespaceID={namespaceID}
                        batchChangeName={match.params.batchChangeName}
                    />
                )}
            />
            <Route
                path={`${match.url}/:batchChangeName`}
                render={({ match, ...props }: RouteComponentProps<{ batchChangeName: string }>) => (
                    <BatchChangeDetailsPage
                        {...outerProps}
                        {...props}
                        namespaceID={namespaceID}
                        batchChangeName={match.params.batchChangeName}
                    />
                )}
            />
            <Route
                path={match.url}
                render={props => (
                    <NamespaceBatchChangeListPage
                        headingElement="h2"
                        {...props}
                        {...outerProps}
                        namespaceID={namespaceID}
                    />
                )}
                exact={true}
            />
        </Switch>
    </div>
))
