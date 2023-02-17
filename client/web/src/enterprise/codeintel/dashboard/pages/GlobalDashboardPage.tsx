import { FunctionComponent } from 'react'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Badge, Container, ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { ExternalRepositoryIcon } from '../../../../site-admin/components/ExternalRepositoryIcon'
import { ConfigurationStateBadge } from '../components/ConfigurationStateBadge'
import { useGlobalCodeIntelStatus } from '../hooks/useGlobalCodeIntelStatus'

export const GlobalDashboardPage: FunctionComponent<{}> = () => {
    const { data, loading, error } = useGlobalCodeIntelStatus({ variables: {} })

    return loading ? (
        <LoadingSpinner />
    ) : error ? (
        <ErrorAlert prefix="Failed to load code intelligence summary" error={error} />
    ) : data ? (
        <>
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code intelligence summary</>,
                    },
                ]}
                className="mb-3"
            />

            <Container className="mb-2">
                <div>
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-success">
                            {data.codeIntelSummary.numRepositoriesWithCodeIntelligence}
                        </span>
                        <span className="text-muted ml-1">Repositories with precise code intelligence</span>
                    </div>
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-danger">
                            {data.codeIntelSummary.repositoriesWithErrors?.nodes.length}
                        </span>
                        <span className="text-muted ml-1">Repositories with problems</span>
                    </div>
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-muted">
                            {data.codeIntelSummary.repositoriesWithConfiguration?.nodes.length}
                        </span>
                        <span className="text-muted ml-1">Configurable repositories</span>
                    </div>
                </div>

                {(data.codeIntelSummary.repositoriesWithErrors?.nodes.length || 0) > 0 && (
                    <div className="mt-2">
                        <span>Repositories with failures:</span>

                        <ul className="list-group p2">
                            {data.codeIntelSummary.repositoriesWithErrors?.nodes.map(({ repository, count }) => (
                                <li key={repository.name} className="list-group-item">
                                    {repository.externalRepository && (
                                        <ExternalRepositoryIcon
                                            externalRepo={{
                                                serviceID: repository.externalRepository.serviceID,
                                                serviceType: repository.externalRepository.serviceType,
                                            }}
                                        />
                                    )}
                                    <RepoLink
                                        repoName={repository.name}
                                        to={`${repository.url}/-/code-graph/dashboard`}
                                    />

                                    <Badge variant="danger" className="ml-2" small={true} pill={true}>
                                        {count}
                                    </Badge>
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                {(data.codeIntelSummary.repositoriesWithConfiguration?.nodes.length || 0) > 0 && (
                    <div className="mt-2">
                        <span>Configurable repositories:</span>

                        <ul className="list-group p2">
                            {data.codeIntelSummary.repositoriesWithConfiguration?.nodes.map(
                                ({ repository, indexers }) => (
                                    <li key={repository.name} className="list-group-item">
                                        {repository.externalRepository && (
                                            <ExternalRepositoryIcon
                                                externalRepo={{
                                                    serviceID: repository.externalRepository.serviceID,
                                                    serviceType: repository.externalRepository.serviceType,
                                                }}
                                            />
                                        )}
                                        <RepoLink
                                            repoName={repository.name}
                                            to={`${repository.url}/-/code-graph/dashboard`}
                                        />

                                        {indexers.map(
                                            indexer =>
                                                indexer.indexer && (
                                                    <span className="ml-2" key={indexer.indexer.key}>
                                                        <ConfigurationStateBadge indexer={indexer.indexer} />

                                                        <Badge variant="info" className="ml-2" small={true} pill={true}>
                                                            {indexer.count}
                                                        </Badge>
                                                    </span>
                                                )
                                        )}
                                    </li>
                                )
                            )}
                        </ul>
                    </div>
                )}
            </Container>
        </>
    ) : (
        <></>
    )
}
