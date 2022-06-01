import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { Redirect, RouteComponentProps } from 'react-router'
import { takeWhile } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { LSIFIndexState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, LoadingSpinner, useObservable, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifIndexFields } from '../../../../graphql-operations'
import { CodeIntelStateBanner, CodeIntelStateBannerProps } from '../../shared/components/CodeIntelStateBanner'
import { CodeIntelAssociatedUpload } from '../components/CodeIntelAssociatedUpload'
import { CodeIntelDeleteIndex } from '../components/CodeIntelDeleteIndex'
import { CodeIntelIndexMeta } from '../components/CodeIntelIndexMeta'
import { CodeIntelIndexTimeline } from '../components/CodeIntelIndexTimeline'
import { queryLisfIndex as defaultQueryLsifIndex } from '../hooks/queryLisfIndex'
import { useDeleteLsifIndex } from '../hooks/useDeleteLsifIndex'

export interface CodeIntelIndexPageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    queryLisfIndex?: typeof defaultQueryLsifIndex
    now?: () => Date
}

const variantByState = new Map<LSIFIndexState, CodeIntelStateBannerProps['variant']>([
    [LSIFIndexState.COMPLETED, 'success'],
    [LSIFIndexState.ERRORED, 'danger'],
])

export const CodeIntelIndexPage: FunctionComponent<React.PropsWithChildren<CodeIntelIndexPageProps>> = ({
    match: {
        params: { id },
    },
    authenticatedUser,
    queryLisfIndex = defaultQueryLsifIndex,
    telemetryService,
    now,
    history,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndex'), [telemetryService])

    const apolloClient = useApolloClient()
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()
    const { handleDeleteLsifIndex, deleteError } = useDeleteLsifIndex()

    useEffect(() => {
        if (deleteError) {
            setDeletionOrError(deleteError)
        }
    }, [deleteError])

    const indexOrError = useObservable(
        useMemo(() => queryLisfIndex(id, apolloClient).pipe(takeWhile(shouldReload, true)), [
            id,
            queryLisfIndex,
            apolloClient,
        ])
    )

    const deleteIndex = useCallback(async (): Promise<void> => {
        if (!indexOrError || isErrorLike(indexOrError)) {
            return
        }

        const autoIndexCommit = indexOrError.inputCommit.slice(0, 7)
        if (!window.confirm(`Delete auto-index record for commit ${autoIndexCommit}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await handleDeleteLsifIndex({
                variables: { id },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            })
            setDeletionOrError('deleted')
            history.push({
                state: {
                    modal: 'SUCCESS',
                    message: `Auto-index record for commit ${autoIndexCommit} has been deleted.`,
                },
            })
        } catch (error) {
            setDeletionOrError(error)
            history.push({
                state: {
                    modal: 'ERROR',
                    message: `There was an error while saving auto-index record for commit: ${autoIndexCommit}.`,
                },
            })
        }
    }, [id, indexOrError, handleDeleteLsifIndex, history])

    return deletionOrError === 'deleted' ? (
        <Redirect to="." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF index record" error={deletionOrError} />
    ) : (
        <div className="site-admin-lsif-index-page w-100">
            <PageTitle title="Auto-indexing jobs" />
            {isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading LSIF index" error={indexOrError} />
            ) : !indexOrError ? (
                <LoadingSpinner />
            ) : (
                <>
                    <PageHeader
                        headingElement="h2"
                        path={[
                            {
                                text: `Auto-index record for ${indexOrError.projectRoot?.repository.name || ''}@${
                                    indexOrError.projectRoot
                                        ? indexOrError.projectRoot.commit.abbreviatedOID
                                        : indexOrError.inputCommit.slice(0, 7)
                                }`,
                            },
                        ]}
                        className="mb-3"
                    />

                    <Container>
                        <CodeIntelIndexMeta node={indexOrError} now={now} />
                    </Container>

                    <Container className="mt-2">
                        <CodeIntelStateBanner
                            state={indexOrError.state}
                            placeInQueue={indexOrError.placeInQueue}
                            failure={indexOrError.failure}
                            typeName="index"
                            pluralTypeName="indexes"
                            variant={variantByState.get(indexOrError.state)}
                        />
                    </Container>

                    {authenticatedUser?.siteAdmin && (
                        <Container className="mt-2">
                            <CodeIntelDeleteIndex deleteIndex={deleteIndex} deletionOrError={deletionOrError} />
                        </Container>
                    )}

                    <Container className="mt-2">
                        <Typography.H3>Timeline</Typography.H3>
                        <CodeIntelIndexTimeline index={indexOrError} now={now} className="mb-3" />
                        <CodeIntelAssociatedUpload node={indexOrError} now={now} />
                    </Container>
                </>
            )}
        </div>
    )
}

const terminalStates = new Set([LSIFIndexState.COMPLETED, LSIFIndexState.ERRORED])

function shouldReload(index: LsifIndexFields | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}
