import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import { Redirect, RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { takeWhile } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, PageHeader, LoadingSpinner, useObservable, Icon, H3 } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { Collapsible } from '../../../../components/Collapsible'
import {
    Connection,
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifUploadFields, LsifUploadConnectionFields } from '../../../../graphql-operations'
import { CodeIntelStateBanner, CodeIntelStateBannerProps } from '../../shared/components/CodeIntelStateBanner'
import { CodeIntelAssociatedIndex } from '../components/CodeIntelAssociatedIndex'
import { CodeIntelDeleteUpload } from '../components/CodeIntelDeleteUpload'
import { CodeIntelUploadMeta } from '../components/CodeIntelUploadMeta'
import { CodeIntelUploadTimeline } from '../components/CodeIntelUploadTimeline'
import { DependencyOrDependentNode } from '../components/DependencyOrDependentNode'
import { EmptyDependencies } from '../components/EmptyDependencies'
import { EmptyDependents } from '../components/EmptyDependents'
import { EmptyUploadRetentionMatchStatus } from '../components/EmptyUploadRetentionStatusNode'
import { RetentionMatchNode } from '../components/UploadRetentionStatusNode'
import { queryLisfUploadFields as defaultQueryLisfUploadFields } from '../hooks/queryLisfUploadFields'
import { queryLsifUploadsList as defaultQueryLsifUploadsList } from '../hooks/queryLsifUploadsList'
import {
    queryUploadRetentionMatches as defaultQueryRetentionMatches,
    NormalizedUploadRetentionMatch,
} from '../hooks/queryUploadRetentionMatches'
import { useDeleteLsifUpload } from '../hooks/useDeleteLsifUpload'

import styles from './CodeIntelUploadPage.module.scss'

export interface CodeIntelUploadPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    queryLisfUploadFields?: typeof defaultQueryLisfUploadFields
    queryLsifUploadsList?: typeof defaultQueryLsifUploadsList
    queryRetentionMatches?: typeof defaultQueryRetentionMatches
    now?: () => Date
}

const variantByState = new Map<LSIFUploadState, CodeIntelStateBannerProps['variant']>([
    [LSIFUploadState.COMPLETED, 'success'],
    [LSIFUploadState.ERRORED, 'danger'],
])

enum DependencyGraphState {
    ShowDependencies,
    ShowDependents,
}

enum RetentionPolicyMatcherState {
    ShowMatchingOnly,
    ShowAll,
}

export const CodeIntelUploadPage: FunctionComponent<React.PropsWithChildren<CodeIntelUploadPageProps>> = ({
    match: {
        params: { id },
    },
    authenticatedUser,
    queryLisfUploadFields = defaultQueryLisfUploadFields,
    queryLsifUploadsList = defaultQueryLsifUploadsList,
    queryRetentionMatches = defaultQueryRetentionMatches,
    telemetryService,
    now,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUpload'), [telemetryService])

    const apolloClient = useApolloClient()
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()
    const [dependencyGraphState, setDependencyGraphState] = useState(DependencyGraphState.ShowDependencies)
    const [retentionPolicyMatcherState, setRetentionPolicyMatcherState] = useState(RetentionPolicyMatcherState.ShowAll)
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

    const queryRetentionPoliciesCallback = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<Connection<NormalizedUploadRetentionMatch>> => {
            if (uploadOrError && !isErrorLike(uploadOrError)) {
                return queryRetentionMatches(apolloClient, id, {
                    matchesOnly: retentionPolicyMatcherState === RetentionPolicyMatcherState.ShowMatchingOnly,
                    ...args,
                })
            }

            throw new Error('unreachable: queryRetentionPolicies referenced with invalid upload')
        },
        [uploadOrError, apolloClient, id, queryRetentionMatches, retentionPolicyMatcherState]
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
                            variant={variantByState.get(uploadOrError.state)}
                        />
                        {uploadOrError.isLatestForRepo && (
                            <div>
                                <Icon role="img" aria-hidden={true} as={InformationOutlineIcon} /> This upload can
                                answer queries for the tip of the default branch and are targets of cross-repository
                                find reference operations.
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
                        <H3>Timeline</H3>
                        <CodeIntelUploadTimeline now={now} upload={uploadOrError} className="mb-3" />
                    </Container>

                    {(uploadOrError.state === LSIFUploadState.COMPLETED ||
                        uploadOrError.state === LSIFUploadState.DELETING) && (
                        <>
                            <Container className="mt-2">
                                <Collapsible
                                    title={
                                        dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                            <H3 className="mb-0">Dependencies</H3>
                                        ) : (
                                            <H3 className="mb-0">Dependents</H3>
                                        )
                                    }
                                    titleAtStart={true}
                                >
                                    {dependencyGraphState === DependencyGraphState.ShowDependencies ? (
                                        <>
                                            <Button
                                                type="button"
                                                className="float-right p-0 mb-2"
                                                variant="link"
                                                onClick={() =>
                                                    setDependencyGraphState(DependencyGraphState.ShowDependents)
                                                }
                                            >
                                                Show dependents
                                            </Button>
                                            <FilteredConnection
                                                listComponent="div"
                                                listClassName={classNames(styles.grid, 'mb-3')}
                                                inputClassName="w-auto"
                                                noun="dependency"
                                                pluralNoun="dependencies"
                                                nodeComponent={DependencyOrDependentNode}
                                                queryConnection={queryDependencies}
                                                history={history}
                                                location={props.location}
                                                cursorPaging={true}
                                                useURLQuery={false}
                                                emptyElement={<EmptyDependencies />}
                                            />
                                        </>
                                    ) : (
                                        <>
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
                                            <FilteredConnection
                                                listComponent="div"
                                                listClassName={classNames(styles.grid, 'mb-3')}
                                                inputClassName="w-auto"
                                                noun="dependent"
                                                pluralNoun="dependents"
                                                nodeComponent={DependencyOrDependentNode}
                                                queryConnection={queryDependents}
                                                history={history}
                                                location={props.location}
                                                cursorPaging={true}
                                                useURLQuery={false}
                                                emptyElement={<EmptyDependents />}
                                            />
                                        </>
                                    )}
                                </Collapsible>
                            </Container>

                            <Container className="mt-2">
                                <Collapsible title={<H3 className="mb-0">Retention overview</H3>} titleAtStart={true}>
                                    {retentionPolicyMatcherState === RetentionPolicyMatcherState.ShowAll ? (
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() =>
                                                setRetentionPolicyMatcherState(
                                                    RetentionPolicyMatcherState.ShowMatchingOnly
                                                )
                                            }
                                        >
                                            Show matching only
                                        </Button>
                                    ) : (
                                        <Button
                                            type="button"
                                            className="float-right p-0 mb-2"
                                            variant="link"
                                            onClick={() =>
                                                setRetentionPolicyMatcherState(RetentionPolicyMatcherState.ShowAll)
                                            }
                                        >
                                            Show all
                                        </Button>
                                    )}
                                    <FilteredConnection
                                        listComponent="div"
                                        listClassName={classNames(styles.grid, 'mb-3')}
                                        inputClassName="w-auto"
                                        noun="match"
                                        pluralNoun="matches"
                                        nodeComponent={RetentionMatchNode}
                                        queryConnection={queryRetentionPoliciesCallback}
                                        history={history}
                                        location={props.location}
                                        cursorPaging={true}
                                        useURLQuery={false}
                                        emptyElement={<EmptyUploadRetentionMatchStatus />}
                                    />
                                </Collapsible>
                            </Container>
                        </>
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
