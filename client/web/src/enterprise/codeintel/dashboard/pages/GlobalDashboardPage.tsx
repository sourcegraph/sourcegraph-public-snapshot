import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Badge, Container, ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'
import { FunctionComponent } from 'react'
import { ExternalRepositoryIcon } from '../../../../site-admin/components/ExternalRepositoryIcon'
import { useGlobalCodeIntelStatus } from '../hooks/useGlobalCodeIntelStatus'

export interface GlobalDashboardPageProps {
    // TODO
}

export const GlobalDashboardPage: FunctionComponent<GlobalDashboardPageProps> = ({}) => {
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
                <Container className="mb-2">
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-success">
                            {data.codeIntelSummary.numRepositoriesWithCodeIntelligence}
                        </span>
                        <span className="text-muted ml-1">Repositories with precise code intelligence</span>
                    </div>
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-danger">
                            {data.codeIntelSummary.repositoriesWithErrors.length}
                        </span>
                        <span className="text-muted ml-1">Repositories with problems</span>
                    </div>
                    <div className="d-inline p-2 m-2 b-2">
                        <span className="d-inline text-muted">
                            {data.codeIntelSummary.repositoriesWithConfiguration.length}
                        </span>
                        <span className="text-muted ml-1">Configurable repositories</span>
                    </div>
                </Container>

                {data.codeIntelSummary.repositoriesWithErrors.length > 0 && (
                    <Container className="mb-2">
                        <span>Repositories with failures:</span>

                        <ul className="list-group p2">
                            {data.codeIntelSummary.repositoriesWithErrors.map(({ repository, count }) => (
                                <li key={repository.name} className="list-group-item">
                                    {repository.externalRepository && (
                                        <ExternalRepositoryIcon externalRepo={repository.externalRepository} />
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
                    </Container>
                )}

                {data.codeIntelSummary.repositoriesWithConfiguration.length > 0 && (
                    <div className="mt-2 b-2">
                        <Container className="mb-2">
                            <span>Configurable repositories:</span>

                            <ul className="list-group p2">
                                {data.codeIntelSummary.repositoriesWithConfiguration.map(({ repository, indexers }) => (
                                    <li key={repository.name} className="list-group-item">
                                        {repository.externalRepository && (
                                            <ExternalRepositoryIcon externalRepo={repository.externalRepository} />
                                        )}
                                        <RepoLink
                                            repoName={repository.name}
                                            to={`${repository.url}/-/code-graph/dashboard`}
                                        />

                                        {indexers.map(indexer => (
                                            <span className="ml-2">
                                                {indexer.indexer?.key}
                                                <Badge variant="info" className="ml-2" small={true} pill={true}>
                                                    {indexer.count}
                                                </Badge>
                                            </span>
                                        ))}
                                    </li>
                                ))}
                            </ul>
                        </Container>
                    </div>
                )}
            </Container>
        </>
    ) : (
        <></>
    )
}
