import { FunctionComponent, useEffect, useState, useCallback } from 'react'

import { mdiDelete } from '@mdi/js'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { Redirect, RouteComponentProps } from 'react-router'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Text, Button, Code, Container, H2, H3, Icon, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../components/HeroPage'
import { PageTitle } from '../../../../components/PageTitle'
import { Timestamp } from '../../../../components/time/Timestamp'
import {
    DeleteLockfileIndexResult,
    DeleteLockfileIndexVariables,
    LockfileIndexResult,
    LockfileIndexVariables,
} from '../../../../graphql-operations'

import { DELETE_LOCKFILE_INDEX, LOCKFILE_INDEX } from './queries'

export interface CodeIntelLockfilePageProps extends RouteComponentProps<{ id: string }>, TelemetryProps {}

export const CodeIntelLockfilePage: FunctionComponent<React.PropsWithChildren<CodeIntelLockfilePageProps>> = ({
    match: {
        params: { id },
    },
    telemetryService,
}) => {
    useEffect(() => telemetryService.logPageView('CodeIntelLockfile'), [telemetryService])

    const { data, loading, error } = useQuery<LockfileIndexResult, LockfileIndexVariables>(LOCKFILE_INDEX, {
        variables: { id },
        fetchPolicy: 'cache-first',
    })

    const [deleted, setDeleted] = useState(false)
    const [isDeleting, setDeleting] = useState(false)
    const [deleteError, setDeleteError] = useState<string>()
    const [deleteLockfileIndex] = useMutation<DeleteLockfileIndexResult, DeleteLockfileIndexVariables>(
        DELETE_LOCKFILE_INDEX
    )

    const deleteIndex = useCallback(() => {
        if (!window.confirm('Delete lockfile index?')) {
            return
        }

        setDeleting(true)
        setDeleteError(undefined)
        deleteLockfileIndex({ variables: { id } })
            .then(() => {
                setDeleting(false)
                setDeleteError(undefined)
                setDeleted(true)
            })
            .catch((error: Error) => {
                setDeleteError(error.message)
                setDeleting(false)
            })
    }, [deleteLockfileIndex, id])

    if (deleted) {
        return <Redirect to="./" />
    }

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <div className="w-100 text-center">
                <Icon aria-label="Loading" className="m-2" as={LoadingSpinner} />
            </div>
        )
    }
    // If we received an error before we successfully received any data
    if (error && !data) {
        throw new Error(error.message)
    }
    // If there weren't any errors and we just didn't receive any data
    if (!data?.node || data.node.__typename !== 'LockfileIndex') {
        return <HeroPage icon={AlertCircleIcon} title="Lockfile index not found" />
    }

    const node = data.node
    const title = `Index of ${node.lockfile} in ${node.repository.name}@${node.commit.abbreviatedOID}`

    return (
        <div className="code-intel-lockfiles">
            <PageTitle title={title} />
            <PageHeader
                headingElement="h1"
                path={[{ text: 'Lockfile index' }]}
                description="Lockfile index created by lockfile-indexing"
                className="mb-3"
            />

            <Container>
                <H2>Overview</H2>
                <div className="m-0">
                    <H3 className="m-0 d-block d-md-inline">
                        <Link to={node.repository.url}>{node.repository.name}</Link>
                    </H3>
                </div>

                <div>
                    <span className="mr-2 d-block d-mdinline-block">
                        Lockfile <Code>{node.lockfile}</Code> indexed at commit{' '}
                        <Link to={node.commit.url}>
                            <Code>{node.commit.abbreviatedOID}</Code>
                        </Link>
                        . Dependency graph fidelity: {node.fidelity}.
                    </span>

                    <small className="text-mute">
                        Indexed <Timestamp date={node.createdAt} />.{' '}
                        {node.createdAt !== node.updatedAt && (
                            <>
                                Updated <Timestamp date={node.updatedAt} />{' '}
                            </>
                        )}
                        .
                    </small>
                </div>
            </Container>

            <Container>
                <H2>Deletion</H2>
                <Text>
                    Deleting this lockfile index will make dependency search unavailable for the repository at the given
                    commit and lockfile. Deletion will not cause packages added to the instance through
                    lockfile-indexing to be deleted.
                </Text>
                <Button
                    type="button"
                    variant="danger"
                    onClick={deleteIndex}
                    disabled={isDeleting || deleteError !== undefined}
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete lockfile index
                </Button>
            </Container>
        </div>
    )
}
