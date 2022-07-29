import React, { useEffect, useState, useCallback, useMemo } from 'react'

import * as H from 'history'
import { parse as parseJSONC } from 'jsonc-parser'
import { useHistory } from 'react-router'
import { Observable, Subject } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, hasProperty } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, H2, H3, Badge, Container } from '@sourcegraph/wildcard'

import {
    ExternalServiceFields,
    Scalars,
    AddExternalServiceInput,
    ExternalServiceSyncJobListFields,
    ExternalServiceSyncJobConnectionFields,
} from '../../graphql-operations'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../FilteredConnection'
import { LoaderButton } from '../LoaderButton'
import { PageTitle } from '../PageTitle'
import { Duration } from '../time/Duration'
import { Timestamp } from '../time/Timestamp'

import {
    isExternalService,
    updateExternalService,
    fetchExternalService as _fetchExternalService,
    useSyncExternalService,
    queryExternalServiceSyncJobs,
} from './backend'
import { ExternalServiceCard } from './ExternalServiceCard'
import { ExternalServiceForm } from './ExternalServiceForm'
import { defaultExternalServices, codeHostExternalServices } from './externalServices'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

interface Props extends TelemetryProps {
    externalServiceID: Scalars['ID']
    isLightTheme: boolean
    history: H.History
    afterUpdateRoute: string

    /** For testing only. */
    fetchExternalService?: typeof _fetchExternalService
    /** For testing only. */
    autoFocusForm?: boolean
}

function isValidURL(url: string): boolean {
    try {
        new URL(url)
        return true
    } catch {
        return false
    }
}

export const ExternalServicePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalServiceID,
    history,
    isLightTheme,
    telemetryService,
    afterUpdateRoute,
    fetchExternalService = _fetchExternalService,
    autoFocusForm,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalService')
    }, [telemetryService])

    const [externalServiceOrError, setExternalServiceOrError] = useState<ExternalServiceFields | ErrorLike>()

    useEffect(() => {
        const subscription = fetchExternalService(externalServiceID)
            .pipe(catchError(error => [asError(error)]))
            .subscribe(result => {
                setExternalServiceOrError(result)
            })
        return () => subscription.unsubscribe()
    }, [externalServiceID, fetchExternalService])

    const onChange = useCallback(
        (input: AddExternalServiceInput) => {
            if (isExternalService(externalServiceOrError)) {
                setExternalServiceOrError({
                    ...externalServiceOrError,
                    ...input,
                    namespace: externalServiceOrError.namespace,
                })
            }
        },
        [externalServiceOrError, setExternalServiceOrError]
    )

    const [isUpdating, setIsUpdating] = useState<boolean | Error>()
    const onSubmit = useCallback(
        async (event?: React.FormEvent<HTMLFormElement>): Promise<void> => {
            if (event) {
                event.preventDefault()
            }
            if (isExternalService(externalServiceOrError)) {
                try {
                    setIsUpdating(true)
                    const updatedService = await updateExternalService({ input: externalServiceOrError })
                    setIsUpdating(false)
                    // If the update was successful, and did not surface a warning, redirect to the
                    // repositories page, adding `?repositoriesUpdated` to the query string so that we display
                    // a banner at the top of the page.
                    if (updatedService.warning) {
                        setExternalServiceOrError(updatedService)
                    } else {
                        history.push(afterUpdateRoute)
                    }
                } catch (error) {
                    setIsUpdating(asError(error))
                }
            }
        },
        [afterUpdateRoute, externalServiceOrError, history]
    )
    let error: ErrorLike | undefined
    if (isErrorLike(isUpdating)) {
        error = isUpdating
    }

    const [
        syncExternalService,
        { error: syncExternalServiceError, loading: syncExternalServiceLoading },
    ] = useSyncExternalService()

    const externalService = (!isErrorLike(externalServiceOrError) && externalServiceOrError) || undefined

    const syncJobUpdates = useMemo(() => new Subject<void>(), [])
    const triggerSync = useCallback(
        () =>
            externalService &&
            syncExternalService({ variables: { id: externalService.id } }).then(() => {
                syncJobUpdates.next()
            }),
        [externalService, syncExternalService, syncJobUpdates]
    )

    let externalServiceCategory = externalService && defaultExternalServices[externalService.kind]
    if (
        externalService &&
        [GQL.ExternalServiceKind.GITHUB, GQL.ExternalServiceKind.GITLAB].includes(externalService.kind)
    ) {
        const parsedConfig: unknown = parseJSONC(externalService.config)
        const url =
            typeof parsedConfig === 'object' &&
            parsedConfig !== null &&
            hasProperty('url')(parsedConfig) &&
            typeof parsedConfig.url === 'string' &&
            isValidURL(parsedConfig.url)
                ? new URL(parsedConfig.url)
                : undefined
        // We have no way of finding out whether a externalservice of kind GITHUB is GitHub.com or GitHub enterprise, so we need to guess based on the URL.
        if (externalService.kind === GQL.ExternalServiceKind.GITHUB && url?.hostname !== 'github.com') {
            externalServiceCategory = codeHostExternalServices.ghe
        }
        // We have no way of finding out whether a externalservice of kind GITLAB is Gitlab.com or Gitlab self-hosted, so we need to guess based on the URL.
        if (externalService.kind === GQL.ExternalServiceKind.GITLAB && url?.hostname !== 'gitlab.com') {
            externalServiceCategory = codeHostExternalServices.gitlab
        }
    }

    return (
        <div>
            {externalService ? (
                <PageTitle title={`External service - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="External service" />
            )}
            <H2>Update code host connection</H2>
            {externalServiceOrError === undefined && <LoadingSpinner />}
            {isErrorLike(externalServiceOrError) && <ErrorAlert className="mb-3" error={externalServiceOrError} />}

            {externalService && (
                <Container className="mb-3">
                    {externalServiceCategory && (
                        <div className="mb-3">
                            <ExternalServiceCard {...externalServiceCategory} namespace={externalService?.namespace} />
                        </div>
                    )}
                    {externalServiceCategory && (
                        <ExternalServiceForm
                            input={{ ...externalService, namespace: externalService.namespace?.id ?? null }}
                            editorActions={externalServiceCategory.editorActions}
                            jsonSchema={externalServiceCategory.jsonSchema}
                            error={error}
                            warning={externalService.warning}
                            mode="edit"
                            loading={isUpdating === true}
                            onSubmit={onSubmit}
                            onChange={onChange}
                            history={history}
                            isLightTheme={isLightTheme}
                            telemetryService={telemetryService}
                            autoFocus={autoFocusForm}
                        />
                    )}
                    <LoaderButton
                        label="Trigger manual sync"
                        alwaysShowLabel={true}
                        variant="secondary"
                        onClick={triggerSync}
                        loading={syncExternalServiceLoading}
                        disabled={syncExternalServiceLoading}
                    />
                    {syncExternalServiceError && <ErrorAlert error={syncExternalServiceError} />}
                    <ExternalServiceWebhook externalService={externalService} className="mt-3" />
                    <ExternalServiceSyncJobsList externalServiceID={externalService.id} updates={syncJobUpdates} />
                </Container>
            )}
        </div>
    )
}

interface ExternalServiceSyncJobsListProps {
    externalServiceID: Scalars['ID']
    updates: Observable<void>
}

const ExternalServiceSyncJobsList: React.FunctionComponent<ExternalServiceSyncJobsListProps> = ({
    externalServiceID,
    updates,
}) => {
    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryExternalServiceSyncJobs({
                first: args.first ?? null,
                externalService: externalServiceID,
            }),
        [externalServiceID]
    )

    const history = useHistory()

    return (
        <>
            <H3 className="mt-3">Recent sync jobs</H3>
            <FilteredConnection<
                ExternalServiceSyncJobListFields,
                Omit<ExternalServiceSyncJobNodeProps, 'node'>,
                {},
                ExternalServiceSyncJobConnectionFields
            >
                className="mb-0"
                listClassName="list-group list-group-flush mb-0"
                noun="sync job"
                pluralNoun="sync jobs"
                queryConnection={queryConnection}
                nodeComponent={ExternalServiceSyncJobNode}
                nodeComponentProps={{}}
                hideSearch={true}
                noSummaryIfAllNodesVisible={true}
                history={history}
                updates={updates}
                location={history.location}
            />
        </>
    )
}

interface ExternalServiceSyncJobNodeProps {
    node: ExternalServiceSyncJobListFields
}

const ExternalServiceSyncJobNode: React.FunctionComponent<ExternalServiceSyncJobNodeProps> = ({ node }) => (
    <li className="list-group-item py-3">
        <div className="d-flex align-items-center justify-content-between">
            <div className="flex-shrink-0 mr-2">
                <Badge>{node.state}</Badge>
            </div>
            <div className="flex-shrink-0">
                {node.startedAt && (
                    <>
                        {node.finishedAt === null && <>Running since </>}
                        {node.finishedAt !== null && <>Ran for </>}
                        <Duration
                            start={node.startedAt}
                            end={node.finishedAt ?? undefined}
                            stableWidth={false}
                            className="d-inline"
                        />
                    </>
                )}
            </div>
            <div className="text-right flex-grow-1">
                <div>
                    {node.startedAt === null && 'Not started yet'}
                    {node.startedAt !== null && (
                        <>
                            Started <Timestamp date={node.startedAt} />
                        </>
                    )}
                </div>
                <div>
                    {node.finishedAt === null && 'Not finished yet'}
                    {node.finishedAt !== null && (
                        <>
                            Finished <Timestamp date={node.finishedAt} />
                        </>
                    )}
                </div>
            </div>
        </div>
        {node.failureMessage && <ErrorAlert error={node.failureMessage} className="mt-2 mb-0" />}
    </li>
)
