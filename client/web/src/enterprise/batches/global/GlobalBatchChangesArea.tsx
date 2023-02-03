import React from 'react'

import { Switch, Route, RouteComponentProps, useParams } from 'react-router-dom'
import { Params } from 'react-router-dom-v5-compat'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { NotFoundPage } from '../../../components/HeroPage'
import type { BatchChangeClosePageProps } from '../close/BatchChangeClosePage'
import type { CreateBatchChangePageProps } from '../create/CreateBatchChangePage'
import type { BatchChangeDetailsPageProps } from '../detail/BatchChangeDetailsPage'
import { TabName } from '../detail/BatchChangeDetailsTabs'
import type { BatchChangeListPageProps, NamespaceBatchChangeListPageProps } from '../list/BatchChangeListPage'
import type { BatchChangePreviewPageProps } from '../preview/BatchChangePreviewPage'

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

interface Props extends ThemeProps, TelemetryProps, SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

/**
 * The global batch changes area.
 */
export const GlobalBatchChangesArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    isSourcegraphDotCom,
    ...props
}) => (
    <div className="w-100">
        <Switch>
            <Route path="/batch-changes" exact={true}>
                <BatchChangeListPage
                    headingElement="h1"
                    canCreate={Boolean(authenticatedUser) && !isSourcegraphDotCom}
                    authenticatedUser={authenticatedUser}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                    {...props}
                />
            </Route>
            {!isSourcegraphDotCom && (
                <Route path="/batch-changes/create" exact={true}>
                    <AuthenticatedCreateBatchChangePage
                        {...props}
                        headingElement="h1"
                        authenticatedUser={authenticatedUser}
                    />
                </Route>
            )}
            <Route render={() => <NotFoundPage pageType="batch changes" />} key="hardcoded-key" />
        </Switch>
    </div>
)

const AuthenticatedCreateBatchChangePage = withAuthenticatedUser<
    CreateBatchChangePageProps & { authenticatedUser: AuthenticatedUser }
>(props => <CreateBatchChangePage {...props} authenticatedUser={props.authenticatedUser} />)

export interface NamespaceBatchChangesAreaProps extends RouteComponentProps, Props {
    namespaceID: Scalars['ID']
}

export const NamespaceBatchChangesArea = withAuthenticatedUser<
    NamespaceBatchChangesAreaProps & { authenticatedUser: AuthenticatedUser }
>(({ match, namespaceID, ...outerProps }) => (
    <div className="pb-3">
        <Switch>
            <Route path={`${match.url}/apply/:specID`}>
                <ExtractParams<{ specID: string }>
                    render={params => <BatchChangePreviewPage {...outerProps} batchSpecID={params.specID!} />}
                />
            </Route>

            <Route path={`${match.url}/:batchChangeName/close`}>
                <ExtractParams<{ batchChangeName: string }>
                    render={params => (
                        <BatchChangeClosePage
                            {...outerProps}
                            namespaceID={namespaceID}
                            batchChangeName={params.batchChangeName!}
                        />
                    )}
                />
            </Route>
            <Route path={`${match.url}/:batchChangeName/executions`}>
                <ExtractParams<{ batchChangeName: string }>
                    render={params => (
                        <BatchChangeDetailsPage
                            {...outerProps}
                            namespaceID={namespaceID}
                            batchChangeName={params.batchChangeName!}
                            initialTab={TabName.Executions}
                        />
                    )}
                />
            </Route>
            <Route path={`${match.url}/:batchChangeName`}>
                <ExtractParams<{ batchChangeName: string }>
                    render={params => (
                        <BatchChangeDetailsPage
                            {...outerProps}
                            namespaceID={namespaceID}
                            batchChangeName={params.batchChangeName!}
                        />
                    )}
                />
            </Route>
            <Route path={match.url}>
                <NamespaceBatchChangeListPage headingElement="h2" {...outerProps} namespaceID={namespaceID} />
            </Route>
        </Switch>
    </div>
))

interface ExtractParamsProps<P extends { [K in keyof Params]?: string }> {
    render: (params: Readonly<[P] extends [string] ? Params<P> : Partial<P>>) => JSX.Element
}
const ExtractParams = <P extends { [K in keyof Params]?: string }>({ render }: ExtractParamsProps<P>): JSX.Element => {
    // TODO: Replace useParams to V6 API once the above V5 <Switch> can be changed to a V6 <Routes>
    const params = useParams<P>()
    return render(params)
}
