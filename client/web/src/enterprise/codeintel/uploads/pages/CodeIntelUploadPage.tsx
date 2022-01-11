import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { takeWhile } from 'rxjs/operators'

import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import {
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { Button, Container, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { ErrorAlert } from '../../../../components/alerts'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifUploadFields, LsifUploadConnectionFields } from '../../../../graphql-operations'
import { CodeIntelStateBanner } from '../../shared/components/CodeIntelStateBanner'
import { CodeIntelAssociatedIndex } from '../components/CodeIntelAssociatedIndex'
import { CodeIntelDeleteUpload } from '../components/CodeIntelDeleteUpload'
import { CodeIntelUploadMeta } from '../components/CodeIntelUploadMeta'
import { CodeIntelUploadTimeline } from '../components/CodeIntelUploadTimeline'
import { DependencyOrDependentNode } from '../components/DependencyOrDependentNode'
import { EmptyDependencies } from '../components/EmptyDependencies'
import { EmptyDependents } from '../components/EmptyDependents'
import { queryLisfUploadFields as defaultQueryLisfUploadFields } from '../hooks/queryLisfUploadFields'
import { queryLsifUploadsList as defaultQueryLsifUploadsList } from '../hooks/queryLsifUploadsList'
import { useDeleteLsifUpload } from '../hooks/useDeleteLsifUpload'

import styles from './CodeIntelUploadPage.module.scss'

export interface CodeIntelUploadPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    queryLisfUploadFields?: typeof defaultQueryLisfUploadFields
    queryLsifUploadsList?: typeof defaultQueryLsifUploadsList
    now?: () => Date
}

const classNamesByState = new Map([
    [LSIFUploadState.COMPLETED, 'alert-success'],
    [LSIFUploadState.ERRORED, 'alert-danger'],
])

enum DependencyGraphState {
    ShowDependencies,
    ShowDependents,
}

export const CodeIntelUploadPage: FunctionComponent<CodeIntelUploadPageProps> = ({
    match: {
        params: { id },
    },
    authenticatedUser,
    queryLisfUploadFields = defaultQueryLisfUploadFields,
    queryLsifUploadsList = defaultQueryLsifUploadsList,
    telemetryService,
    now,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUpload'), [telemetryService])

    const apolloClient = useApolloClient()
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()
    const [dependencyGraphState, setDependencyGraphState] = useState(DependencyGraphState.ShowDependencies)
    const { handleDeleteLsifUpload, deleteError } = useDeleteLsifUpload()

    useEffect(() => {
        if (deleteError) {
            setDeletionOrError(deleteError)
        }
    }, [deleteError])

    const uploadOrError = useObservable(
        useMemo(() => queryLisfUploadFields(id, apolloClient).pipe(takeWhile(shouldReload, true)), [
            id,
            queryLisfUploadFields,
            apolloClient,
        ])
    )

    const deleteUpload = useCallback(async (): Promise<void> => {
        if (!uploadOrError || isErrorLike(uploadOrError)) {
            return
        }

        let description = `${uploadOrError.inputCommit.slice(0, 7)}`
        if (uploadOrError.inputRoot) {
            description += ` rooted at ${uploadOrError.inputRoot}`
        }

        if (!window.confirm(`Delete upload for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await handleDeleteLsifUpload({
                variables: { id },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            })
            setDeletionOrError('deleted')
            history.push({
                state: {
                    modal: 'SUCCESS',
                    message: `Upload for commit ${description} is deleting.`,
                },
            })
        } catch (error) {
            setDeletionOrError(error)
            history.push({
                state: {
                    modal: 'ERROR',
                    message: `There was an error while deleting upload for commit ${description}.`,
                },
            })
        }
    }, [id, uploadOrError, handleDeleteLsifUpload, history])

    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<LsifUploadConnectionFields> => {
            if (uploadOrError && !isErrorLike(uploadOrError)) {
                return queryLsifUploadsList({ ...args, dependencyOf: uploadOrError.id }, apolloClient)
            }
            throw new Error('unreachable: queryDependencies referenced with invalid upload')
        },
        [uploadOrError, queryLsifUploadsList, apolloClient]
    )

    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (uploadOrError && !isErrorLike(uploadOrError)) {
                return queryLsifUploadsList({ ...args, dependentOf: uploadOrError.id }, apolloClient)
            }

            throw new Error('unreachable: queryDependents referenced with invalid upload')
        },
        [uploadOrError, queryLsifUploadsList, apolloClient]
    )

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-upload-page w-100">
            <PageTitle title="Precise code intelligence uploads" />
            {isErrorLike(uploadOrError) ? (
                <ErrorAlert prefix="Error loading LSIF upload" error={uploadOrError} />
            ) : !uploadOrError ? (
                <LoadingSpinner />
            ) : (
                <>
                    <PageHeader
                        headingElement="h2"
                        path={[
                            {
                                text: `Upload for commit ${uploadOrError.projectRoot?.repository.name || ''}@${
                                    uploadOrError.projectRoot
                                        ? uploadOrError.projectRoot.commit.abbreviatedOID
                                        : uploadOrError.inputCommit.slice(0, 7)
                                }`,
                            },
                        ]}
                        className="mb-3"
                    />

                    <Container>
                        <CodeIntelUploadMeta node={uploadOrError} now={now} />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelStateBanner
                            state={uploadOrError.state}
                            placeInQueue={uploadOrError.placeInQueue}
                            failure={uploadOrError.failure}
                            typeName="upload"
                            pluralTypeName="uploads"
                            className={classNamesByState.get(uploadOrError.state)}
                        />
                        {uploadOrError.isLatestForRepo && (
                            <div>
                                <InformationOutlineIcon className="icon-inline" /> This upload can answer queries for
                                the tip of the default branch and are targets of cross-repository find reference
                                operations.
                            </div>
                        )}
                    </Container>

                    {authenticatedUser?.siteAdmin && (
                        <Container className="mt-2">
                            <CodeIntelDeleteUpload
                                state={uploadOrError.state}
                                deleteUpload={deleteUpload}
                                deletionOrError={deletionOrError}
                            />
                        </Container>
                    )}

                    <Container className="mt-2">
                        <CodeIntelAssociatedIndex node={uploadOrError} now={now} />
                        <h3>Timeline</h3>
                        <CodeIntelUploadTimeline now={now} upload={uploadOrError} className="mb-3" />
                    </Container>

                    {(uploadOrError.state === LSIFUploadState.COMPLETED ||
                        uploadOrError.state === LSIFUploadState.DELETING) && (
                        <Container className="mt-2">
                            <div className="mb-2">
                                {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                    <h3>
                                        Dependencies
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() => setDependencyGraphState(DependencyGraphState.ShowDependents)}
                                        >
                                            Show dependents
                                        </Button>
                                    </h3>
                                ) : (
                                    <h3>
                                        Dependents
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() =>
                                                setDependencyGraphState(DependencyGraphState.ShowDependencies)
                                            }
                                        >
                                            Show dependencies
                                        </Button>
                                    </h3>
                                )}
                            </div>

                            {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName={classNames(styles.grid, 'mb-3')}
                                    noun="dependency"
                                    pluralNoun="dependencies"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependencies}
                                    history={history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependencies />}
                                />
                            ) : (
                                <FilteredConnection
                                    listComponent="div"
                                    listClassName={classNames(styles.grid, 'mb-3')}
                                    noun="dependent"
                                    pluralNoun="dependents"
                                    nodeComponent={DependencyOrDependentNode}
                                    queryConnection={queryDependents}
                                    history={history}
                                    location={props.location}
                                    cursorPaging={true}
                                    emptyElement={<EmptyDependents />}
                                />
                            )}
                        </Container>
                    )}
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFUploadState.COMPLETED, LSIFUploadState.ERRORED, LSIFUploadState.DELETING])

function shouldReload(upload: LsifUploadFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(upload) && !(upload && terminalStates.has(upload.state))
}
