import React, { useCallback, useEffect, useMemo, useContext } from 'react'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import PlusIcon from 'mdi-react/PlusIcon'
import GearIcon from 'mdi-react/GearIcon'
import { Link } from 'react-router-dom'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { PageHeader } from '../../../components/PageHeader'
import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Page } from '../../../components/Page'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { InsightsIcon, InsightsViewGrid, InsightsViewGridProps } from '../components'
import { InsightsApiContext } from '../core/backend/api-provider';

interface InsightsPageProps extends ExtensionsControllerProps, Omit<InsightsViewGridProps, 'views'>, TelemetryProps {}

export const InsightsPage: React.FunctionComponent<InsightsPageProps> = props => {
    const { getInsightCombinedViews } = useContext(InsightsApiContext);

    const views = useObservable(
        useMemo(
            () => getInsightCombinedViews(props.extensionsController?.extHostAPI),
            [props.extensionsController, getInsightCombinedViews]
        )
    )

    useEffect(() => {
        props.telemetryService.logViewEvent('Insights')
    }, [props.telemetryService])

    const logConfigureClick = useCallback(() => {
        props.telemetryService.log('InsightConfigureClick')
    }, [props.telemetryService])

    const logAddMoreClick = useCallback(() => {
        props.telemetryService.log('InsightAddMoreClick')
    }, [props.telemetryService])

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                    path={[{ icon: InsightsIcon, text: 'Code insights' }]}
                    actions={
                        <>
                            <Link
                                to="/extensions?query=category:Insights"
                                onClick={logAddMoreClick}
                                className="btn btn-secondary mr-1"
                            >
                                <PlusIcon className="icon-inline" /> Add more insights
                            </Link>
                            <Link to="/user/settings" onClick={logConfigureClick} className="btn btn-secondary">
                                <GearIcon className="icon-inline" /> Configure insights
                            </Link>
                        </>
                    }
                    className="mb-3"
                />
                {views === undefined ? (
                    <div className="d-flex w-100">
                        <LoadingSpinner className="my-4" />
                    </div>
                ) : (
                    <InsightsViewGrid {...props} views={views} />
                )}
            </Page>
        </div>
    )
}
