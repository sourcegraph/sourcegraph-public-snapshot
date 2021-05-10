import classnames from 'classnames';
import React, { useContext, useEffect, useState } from 'react'
import { useHistory } from 'react-router-dom';
import type { LineChartContent } from 'sourcegraph';

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService';
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors';

import { ErrorAlert } from '../../../../../../components/alerts';
import { ChartViewContent } from '../../../../../../views/ChartViewContent/ChartViewContent';
import { InsightsApiContext } from '../../../../../core/backend/api-provider';
import { DataSeries } from '../../types';

import { DEFAULT_MOCK_CHART_CONTENT } from './live-preview-mock-data';
import styles from './SearchInsightLivePreview.module.scss'

export interface SearchInsightLivePreviewProps {
    className?: string
    repositories: string
    series: DataSeries[],
    stepValue: string,
    step: 'hours' | 'days' | 'weeks' | 'months' | 'years'
}

export const SearchInsightLivePreview: React.FunctionComponent<SearchInsightLivePreviewProps> = props => {
    const { series, repositories, step, stepValue, className } = props;

    const history = useHistory()
    const { getSearchInsightContent } = useContext(InsightsApiContext)
    const [loading, setLoading] = useState<boolean>(false);
    const [dataOrError, setDataOrError] = useState<LineChartContent<any, string> | Error | undefined>(DEFAULT_MOCK_CHART_CONTENT);

    useEffect(() => {
        if (!repositories || series.length === 0) {
            return;
        }

        let hasRequestCanceled = false;

        setLoading(true)

        const liveSettings = {
            series: series.map(line => ({...line, query: line.query.replace(/\\\\/g, '\\') })),
            repositories: repositories.trim().split(/\s*,\s*/),
            step: { [step]: stepValue }
        }

        getSearchInsightContent(liveSettings)
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => setLoading(false))

        return () => {
            hasRequestCanceled = true;
        }
    }, [getSearchInsightContent, series, repositories, step, stepValue])

    return (
        <div className={classnames(styles.livePreview, className)}>

            { (loading || !dataOrError) && <span>Loading...</span>}
            { isErrorLike(dataOrError) && <ErrorAlert className="m-0" error={dataOrError} />}

            { !(loading || isErrorLike(dataOrError) || !dataOrError)  &&
                <ChartViewContent
                    history={history}
                    viewID='search-insight-live-preview'
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    content={dataOrError}
                />
            }
        </div>
    );
}
