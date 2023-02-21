import { useQuery } from '@sourcegraph/http-client'
import { Container, ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { GlobalCodeIntelStatusResult } from '../../../../graphql-operations'
import { globalCodeIntelStatusQuery } from '../backend'

export const GlobalDashboardPage: React.FunctionComponent = () => {
    const { data, error, loading } = useQuery<GlobalCodeIntelStatusResult>(globalCodeIntelStatusQuery, {
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

    if (loading || !data) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert prefix="Failed to load code intelligence summary" error={error} />
    }

    return (
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
            <Container>
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
            </Container>
        </>
    )
}
