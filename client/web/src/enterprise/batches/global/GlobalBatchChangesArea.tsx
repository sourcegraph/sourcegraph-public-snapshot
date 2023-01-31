import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Routes, Route, useParams, Params } from 'react-router-dom-v5-compat'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { HeroPage } from '../../../components/HeroPage'
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
        <Routes>
            <Route
                path=""
                element={
                    <BatchChangeListPage
                        headingElement="h1"
                        canCreate={Boolean(authenticatedUser) && !isSourcegraphDotCom}
                        authenticatedUser={authenticatedUser}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        {...props}
                    />
                }
            />
            {!isSourcegraphDotCom && (
                <Route
                    path="create"
                    element={
                        <AuthenticatedCreateBatchChangePage
                            {...props}
                            headingElement="h1"
                            authenticatedUser={authenticatedUser}
                        />
                    }
                />
            )}
            <Route element={<NotFoundPage />} key="hardcoded-key" />
        </Routes>
    </div>
)

const AuthenticatedCreateBatchChangePage = withAuthenticatedUser<
    CreateBatchChangePageProps & { authenticatedUser: AuthenticatedUser }
>(props => <CreateBatchChangePage {...props} authenticatedUser={props.authenticatedUser} />)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

export interface NamespaceBatchChangesAreaProps extends Props {
    namespaceID: Scalars['ID']
}

export const NamespaceBatchChangesArea = withAuthenticatedUser<
    NamespaceBatchChangesAreaProps & { authenticatedUser: AuthenticatedUser }
>(({ namespaceID, ...outerProps }) => (
    <div className="pb-3">
        <Routes>
            <Route
                path="apply/:specID"
                element={
                    <ExtractParams
                        render={params => <BatchChangePreviewPage {...outerProps} batchSpecID={params.specID!} />}
                    />
                }
            />
            <Route
                path=":batchChangeName/close"
                element={
                    <ExtractParams
                        render={params => (
                            <BatchChangeClosePage
                                {...outerProps}
                                namespaceID={namespaceID}
                                batchChangeName={params.batchChangeName!}
                            />
                        )}
                    />
                }
            />
            <Route
                path=":batchChangeName/executions"
                element={
                    <ExtractParams
                        render={params => (
                            <BatchChangeDetailsPage
                                {...outerProps}
                                namespaceID={namespaceID}
                                batchChangeName={params.batchChangeName!}
                                initialTab={TabName.Executions}
                            />
                        )}
                    />
                }
            />
            <Route
                path=":batchChangeName"
                element={
                    <ExtractParams
                        render={params => (
                            <BatchChangeDetailsPage
                                {...outerProps}
                                namespaceID={namespaceID}
                                batchChangeName={params.batchChangeName!}
                            />
                        )}
                    />
                }
            />
            <Route
                path=""
                element={<NamespaceBatchChangeListPage headingElement="h2" {...outerProps} namespaceID={namespaceID} />}
            />
        </Routes>
    </div>
))

interface ExtractParamsProps {
    render: (params: Readonly<Params<string>>) => JSX.Element
}
const ExtractParams: React.FC<ExtractParamsProps> = ({ render }) => {
    const params = useParams()
    return render(params)
}
