import { FunctionComponent, useEffect } from 'react'

import { formatDistance } from 'date-fns'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../../components/Collapsible'

import { useRankingSummary as defaultUseRankingSummary } from './backend'

export interface CodeIntelRankingPageProps extends TelemetryProps {
    useRankingSummary?: typeof defaultUseRankingSummary
    telemetryService: TelemetryService
}

export const CodeIntelRankingPage: FunctionComponent<CodeIntelRankingPageProps> = ({
    useRankingSummary = defaultUseRankingSummary,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelRankingPage'), [telemetryService])

    const { data, loading, error } = useRankingSummary({})

    if (loading && !data) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert prefix="Failed to load code intelligence summary for repository" error={error} />
    }

    return (
        <>
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Ranking calculation history</>,
                    },
                ]}
                description="View the history of ranking calculation."
                className="mb-3"
            />

            <Container>
                {data &&
                    (data.rankingSummary.length === 0 ? (
                        <>No data.</>
                    ) : (
                        <>
                            <h3>Current ranking calculation</h3>

                            <div className="p-2">
                                <Summary key={data.rankingSummary[0].graphKey} summary={data.rankingSummary[0]} />
                            </div>
                        </>
                    ))}
            </Container>

            {data && data.rankingSummary.length > 1 && (
                <Container>
                    <Collapsible title="Historic ranking calculations" titleAtStart={true} titleClassName="h3">
                        <div className="p-4">
                            {data.rankingSummary.slice(1).map(summary => (
                                <Summary key={summary.graphKey} summary={summary} />
                            ))}
                        </div>
                    </Collapsible>
                </Container>
            )}
        </>
    )
}

interface Summary {
    graphKey: string
    pathMapperProgress: Progress
    referenceMapperProgress: Progress
    reducerProgress: Progress | null
}

interface Progress {
    startedAt: string
    completedAt: string | null
    processed: number
    total: number
}

interface SummaryProps {
    summary: Summary
}

const Summary: FunctionComponent<SummaryProps> = ({ summary }) => (
    <div>
        <strong>Graph key: {summary.graphKey}</strong>

        <ul>
            <li>
                Path mapper: <Progress progress={summary.pathMapperProgress} />
            </li>

            <li>
                Reference mapper: <Progress progress={summary.referenceMapperProgress} />
            </li>

            {summary.reducerProgress && (
                <li>
                    Reducer: <Progress progress={summary.reducerProgress} />
                </li>
            )}
        </ul>
    </div>
)

interface ProgressProps {
    progress: Progress
}

const Progress: FunctionComponent<ProgressProps> = ({ progress }) => (
    <span>
        <span className="d-block">
            {progress.processed === 0 ? (
                <>No records marked for processing</>
            ) : (
                <>
                    {progress.processed} of {progress.total} records processed
                </>
            )}

            <span className="float-right">
                {progress.processed === 0 ? 100 : Math.floor(((progress.processed / progress.total) * 100 * 100) / 100)}
                %
            </span>
        </span>

        <span className="d-block text-muted">
            Started running <Timestamp date={progress.startedAt} />
            {progress.completedAt && (
                <> and ran for {formatDistance(new Date(progress.completedAt), new Date(progress.startedAt))}</>
            )}
            .
        </span>
    </span>
)
