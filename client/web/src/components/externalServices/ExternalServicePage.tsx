import React, { useEffect, useState, useCallback, useMemo, FC } from 'react'

import { mdiCog, mdiConnection, mdiDelete } from '@mdi/js'
import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Subject } from 'rxjs'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Button, Container, ErrorAlert, H2, Icon, Link, PageHeader, Tooltip } from '@sourcegraph/wildcard'

import { Scalars, ExternalServiceResult, ExternalServiceVariables } from '../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../settings/DynamicallyImportedMonacoSettingsEditor'
import { refreshSiteFlags } from '../../site/backend'
import { CreatedByAndUpdatedByInfoByline } from '../Byline/CreatedByAndUpdatedByInfoByline'
import { HeroPage } from '../HeroPage'
import { LoaderButton } from '../LoaderButton'
import { PageTitle } from '../PageTitle'

import {
    useSyncExternalService,
    queryExternalServiceSyncJobs as _queryExternalServiceSyncJobs,
    FETCH_EXTERNAL_SERVICE,
    deleteExternalService,
    useExternalServiceCheckConnectionByIdLazyQuery,
} from './backend'
import { ExternalServiceInformation } from './ExternalServiceInformation'
import { resolveExternalServiceCategory } from './externalServices'
import { ExternalServiceSyncJobsList } from './ExternalServiceSyncJobsList'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

interface Props extends TelemetryProps {
    externalServiceID: Scalars['ID']
    isLightTheme: boolean
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    queryExternalServiceSyncJobs?: typeof _queryExternalServiceSyncJobs
}

const NotFoundPage: FC<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested code host was not found." />
)

export const ExternalServicePage: FC<React.PropsWithChildren<Props>> = props => {
    const {
        externalServiceID,
        isLightTheme,
        history,
        routingPrefix,
        telemetryService,
        afterDeleteRoute,
        externalServicesFromFile,
        allowEditExternalServicesWithFile,
        queryExternalServiceSyncJobs = _queryExternalServiceSyncJobs,
    } = props

    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalService')
    }, [telemetryService])

    const [syncInProgress, setSyncInProgress] = useState<boolean>(false)
    // Callback used in ExternalServiceSyncJobsList to update the state in current component.
    const updateSyncInProgress = useCallback((updatedSyncInProgress: boolean) => {
        setSyncInProgress(updatedSyncInProgress)
    }, [])

    const {
        data: externalServiceData,
        error: fetchError,
        loading: fetchLoading,
    } = useQuery<ExternalServiceResult, ExternalServiceVariables>(FETCH_EXTERNAL_SERVICE, {
        variables: { id: externalServiceID },
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

    const externalService =
        externalServiceData?.node?.__typename === 'ExternalService' ? externalServiceData.node : undefined

    const [numberOfRepos, setNumberOfRepos] = useState<number>(externalService?.repoCount ?? 0)
    // Callback used in ExternalServiceSyncJobsList to update the number of repos in current component.
    const updateNumberOfRepos = useCallback((updatedNumberOfRepos: number) => {
        setNumberOfRepos(updatedNumberOfRepos)
    }, [])

    const [syncExternalService, { error: syncExternalServiceError, loading: syncExternalServiceLoading }] =
        useSyncExternalService()

    const syncJobUpdates = useMemo(() => new Subject<void>(), [])
    const triggerSync = useCallback(
        () =>
            externalService &&
            syncExternalService({ variables: { id: externalService.id } }).then(() => {
                syncJobUpdates.next()
            }),
        [externalService, syncExternalService, syncJobUpdates]
    )

    const externalServiceCategory = resolveExternalServiceCategory(externalService)

    const editingEnabled = allowEditExternalServicesWithFile || !externalServicesFromFile

    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!externalService) {
            return
        }
        if (!window.confirm(`Delete the external service ${externalService.displayName}?`)) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteExternalService(externalService.id)
            setIsDeleting(false)
            // eslint-disable-next-line rxjs/no-ignored-subscription
            refreshSiteFlags().subscribe()
            history.push(afterDeleteRoute)
        } catch (error) {
            setIsDeleting(asError(error))
        }
    }, [afterDeleteRoute, history, externalService])

    // If external service is undefined, we won't use doCheckConnection anyway,
    // that's why it's safe to pass an empty ID to useExternalServiceCheckConnectionByIdLazyQuery
    const [doCheckConnection, { loading, data, error }] = useExternalServiceCheckConnectionByIdLazyQuery(
        externalService?.id || ''
    )

    const checkConnectionNode = data?.node?.__typename === 'ExternalService' ? data.node.checkConnection : null

    let externalServiceAvailabilityStatus
    if (!error && !loading) {
        if (checkConnectionNode?.__typename === 'ExternalServiceAvailable') {
            externalServiceAvailabilityStatus = (
                <Alert className="mt-2" variant="success">
                    Code host is reachable.
                </Alert>
            )
        } else if (checkConnectionNode?.__typename === 'ExternalServiceUnavailable') {
            externalServiceAvailabilityStatus = (
                <ErrorAlert
                    className="mt-2"
                    prefix="Error during code host connection check"
                    error={checkConnectionNode.suspectedReason}
                />
            )
        }
    }

    return (
        <div>
            {externalService ? (
                <PageTitle title={`Code host - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="Code host" />
            )}
            {fetchError !== undefined && !fetchLoading && <ErrorAlert className="mb-3" error={fetchError} />}
            {!fetchLoading && !externalService && !fetchError && <NotFoundPage />}
            {externalService && (
                <Container className="mb-3">
                    <PageHeader
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/external-services', text: 'Code hosts' },
                            { text: externalService.displayName },
                        ]}
                        byline={
                            <CreatedByAndUpdatedByInfoByline
                                createdAt={externalService.createdAt}
                                updatedAt={externalService.updatedAt}
                                noAuthor={true}
                            />
                        }
                        className="mb-3"
                        headingElement="h2"
                        actions={
                            <div className="d-flex align-items-center justify-content-between">
                                <div className="align-self-start">
                                    <Tooltip
                                        content={
                                            externalService.hasConnectionCheck
                                                ? 'Test if code host is reachable from Sourcegraph'
                                                : 'Connection check unavailable'
                                        }
                                    >
                                        <Button
                                            className="test-connection-external-service-button"
                                            variant="secondary"
                                            onClick={() => doCheckConnection()}
                                            disabled={!externalService.hasConnectionCheck || loading}
                                            size="sm"
                                        >
                                            <Icon aria-hidden={true} svgPath={mdiConnection} /> Test connection
                                        </Button>
                                    </Tooltip>
                                </div>
                                {editingEnabled && (
                                    <div className="flex-grow-1 ml-1">
                                        <Tooltip content="Edit code host connection settings">
                                            <Button
                                                className="test-edit-external-service-button"
                                                to={`${routingPrefix}/external-services/${externalService.id}/edit`}
                                                variant="primary"
                                                size="sm"
                                                as={Link}
                                            >
                                                <Icon aria-hidden={true} svgPath={mdiCog} />
                                                {' Edit'}
                                            </Button>
                                        </Tooltip>
                                    </div>
                                )}
                                <div className="flex-shrink-0 ml-1">
                                    <Tooltip content="Delete code host connection">
                                        <Button
                                            aria-label="Delete"
                                            className="test-delete-external-service-button"
                                            onClick={onDelete}
                                            disabled={isDeleting === true}
                                            variant="danger"
                                            size="sm"
                                        >
                                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                                            {' Delete'}
                                        </Button>
                                    </Tooltip>
                                </div>
                            </div>
                        }
                    />
                    {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
                    {externalServiceAvailabilityStatus}
                    <H2>Information</H2>
                    {externalServiceCategory && (
                        <ExternalServiceInformation
                            displayName={externalService.displayName}
                            codeHostID={externalService.id}
                            reposNumber={numberOfRepos === 0 ? externalService.repoCount : numberOfRepos}
                            syncInProgress={syncInProgress}
                            {...externalServiceCategory}
                        />
                    )}
                    <H2>Configuration</H2>
                    {externalServiceCategory && (
                        <DynamicallyImportedMonacoSettingsEditor
                            value={externalService.config}
                            jsonSchema={externalServiceCategory.jsonSchema}
                            canEdit={false}
                            loading={fetchLoading}
                            height={350}
                            readOnly={true}
                            isLightTheme={isLightTheme}
                            history={history}
                            className="test-external-service-editor"
                            telemetryService={telemetryService}
                        />
                    )}
                    <ExternalServiceWebhook externalService={externalService} className="mt-3" />
                    <LoaderButton
                        label="Trigger manual sync"
                        className="mt-3 mb-2 float-right"
                        alwaysShowLabel={true}
                        variant="secondary"
                        onClick={triggerSync}
                        loading={syncExternalServiceLoading}
                        disabled={syncExternalServiceLoading}
                    />
                    {syncExternalServiceError && <ErrorAlert error={syncExternalServiceError} />}
                    <ExternalServiceSyncJobsList
                        queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                        updateSyncInProgress={updateSyncInProgress}
                        updateNumberOfRepos={updateNumberOfRepos}
                        externalServiceID={externalService.id}
                        updates={syncJobUpdates}
                    />
                </Container>
            )}
        </div>
    )
}
