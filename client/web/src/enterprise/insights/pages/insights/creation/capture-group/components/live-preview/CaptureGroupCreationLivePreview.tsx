import React from 'react'

import RefreshIcon from 'mdi-react/RefreshIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, useDeepMemo } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ParentSize } from '../../../../../../../../charts'
import { getSanitizedRepositories } from '../../../../../../components/creation-ui-kit'
import {
    InsightCard,
    InsightCardBanner,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../../components/views'
import { InsightStep } from '../../../search-insight'
import { getSanitizedCaptureQuery } from '../../utils/capture-group-insight-sanitizer'

import { MOCK_CHART_CONTENT } from './constants'
import { useCaptureGroupPreviewContent } from './use-capture-group-preview-content'

import styles from './CaptureGroupCreationLivePreview.module.scss'

interface CaptureGroupCreationLivePreviewProps {
    disabled: boolean
    repositories: string
    query: string
    stepValue: string
    step: InsightStep
    isAllReposMode: boolean
    className?: string
}

export const CaptureGroupCreationLivePreview: React.FunctionComponent<CaptureGroupCreationLivePreviewProps> = props => {
    const { disabled, repositories, query, stepValue, step, isAllReposMode, className } = props

    const settings = useDeepMemo({
        disabled,
        query: getSanitizedCaptureQuery(query.trim()),
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
    })

    const { loading, dataOrError, update } = useCaptureGroupPreviewContent(settings)

    return (
        <aside className={className}>
            <Button variant="icon" disabled={disabled} onClick={update}>
                Live preview <RefreshIcon size="1rem" />
            </Button>

            <InsightCard className={styles.insightCard}>
                {loading ? (
                    <InsightCardLoading>Loading code insight</InsightCardLoading>
                ) : isErrorLike(dataOrError) ? (
                    <ErrorAlert error={dataOrError} />
                ) : (
                    <ParentSize className={styles.chartBlock}>
                        {parent =>
                            dataOrError ? (
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
                                    {...dataOrError}
                                />
                            ) : (
                                <>
                                    <SeriesChart
                                        type={SeriesBasedChartTypes.Line}
                                        width={parent.width}
                                        height={parent.height}
                                        className={styles.chartWithMock}
                                        {...MOCK_CHART_CONTENT}
                                    />
                                    <InsightCardBanner className={styles.disableBanner}>
                                        {isAllReposMode
                                            ? 'Live previews are currently not available for insights running over all repositories.'
                                            : 'The chart preview will be shown here once you have filled out the repositories and series fields.'}
                                    </InsightCardBanner>
                                </>
                            )
                        }
                    </ParentSize>
                )}

                {dataOrError && !isErrorLike(dataOrError) && (
                    <LegendList className="mt-3">
                        {dataOrError.series.map(series => (
                            <LegendItem key={series.dataKey} color={getLineColor(series)} name={series.name} />
                        ))}
                    </LegendList>
                )}
            </InsightCard>
            {isAllReposMode && <p>Previews are only displayed if you individually list up to 50 repositories.</p>}
        </aside>
    )
}
