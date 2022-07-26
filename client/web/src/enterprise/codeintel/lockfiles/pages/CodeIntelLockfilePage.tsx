import { FunctionComponent, useEffect } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Code, Container, H2, H3, Icon, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../components/HeroPage'
import { PageTitle } from '../../../../components/PageTitle'
import { Timestamp } from '../../../../components/time/Timestamp'
import { LockfileIndexResult, LockfileIndexVariables } from '../../../../graphql-operations'

import { LOCKFILE_INDEX } from './queries'

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
        </div>
    )
}
