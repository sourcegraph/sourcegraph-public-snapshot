import { FunctionComponent, useEffect } from 'react'

import { formatDistance, format, parseISO } from 'date-fns'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, LoadingSpinner, PageHeader, H4, H3 } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../../components/Collapsible'

import { useRankingSummary as defaultUseRankingSummary } from './backend'

import styles from './CodeIntelRankingPage.module.scss'

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
                            <H3>Current ranking calculation</H3>

                            <div className="py-3">
                                <Summary key={data.rankingSummary[0].graphKey} summary={data.rankingSummary[0]} />
                            </div>
                        </>
                    ))}
            </Container>

            {data && data.rankingSummary.length > 1 && (
                <Container>
                    <Collapsible title="Historic ranking calculations" titleAtStart={true} titleClassName="h3">
                        <div className="py-3">
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
        <Progress title="Path Aggregation Process" progress={summary.pathMapperProgress} />
        <Progress title="Reference Aggregation Process" progress={summary.referenceMapperProgress} />
        {summary.reducerProgress && <Progress title="Reducing Process" progress={summary.reducerProgress} />}
    </div>
)

interface ProgressProps {
    title: string
    progress: Progress
}

const Progress: FunctionComponent<ProgressProps> = ({ title, progress }) => (
    <div>
        <div className={styles.tableContainer}>
            <H4>{title}</H4>
            <div className={styles.row}>
                <div>Queued records</div>
                <div>
                    {progress.processed === 0 ? (
                        <>Process finished</>
                    ) : (
                        <>
                            {progress.processed} of {progress.total} records processed
                        </>
                    )}
                </div>
            </div>
            <div className={styles.row}>
                <div>Started</div>
                <div>
                    {format(parseISO(progress.startedAt), 'MMM L y h:mm:ss a')} (
                    <Timestamp date={progress.startedAt} />){' '}
                </div>
            </div>
            <div className={styles.row}>
                <div>Completed</div>
                <div>
                    {progress.completedAt ? (
                        <>
                            {format(parseISO(progress.completedAt), 'MMM L y h:mm:ss a')} (
                            <Timestamp date={progress.completedAt} />){' '}
                        </>
                    ) : (
                        '-'
                    )}
                </div>
            </div>
            <div className={styles.row}>
                <div>Duration</div>
                <div>
                    {progress.completedAt && (
                        <> Ran for {formatDistance(new Date(progress.completedAt), new Date(progress.startedAt))}</>
                    )}
                </div>
            </div>

            <div className={styles.row}>
                <div>Progress</div>
                <div>
                    {progress.processed === 0
                        ? 100
                        : Math.floor(((progress.processed / progress.total) * 100 * 100) / 100)}
                    %
                </div>
            </div>
        </div>
    </div>
)
