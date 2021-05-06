import { uniqBy } from 'lodash'
import GearIcon from 'mdi-react/GearIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useMemo, useContext } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Page } from '../../../components/Page'
import { PageHeader } from '../../../components/PageHeader'
import { InsightsIcon, InsightsViewGrid, InsightsViewGridProps } from '../../components'
import { InsightsApiContext } from '../../core/backend/api-provider'
import { FeedbackBadge } from '../../components/FeedbackBadge'
import { Page } from '../../components/Page'
import { InsightsIcon, InsightsViewGrid, InsightsViewGridProps } from '../components'
import { InsightsApiContext } from '../core/backend/api-provider'

export interface InsightsPageProps
    extends ExtensionsControllerProps,
        Omit<InsightsViewGridProps, 'views'>,
        TelemetryProps {
    isCreationUIEnabled: boolean
}

export const InsightsPage: React.FunctionComponent<InsightsPageProps> = props => {
    const { isCreationUIEnabled } = props
    const { getInsightCombinedViews } = useContext(InsightsApiContext)

    const views = useObservable(
        useMemo(() => getInsightCombinedViews(props.extensionsController?.extHostAPI), [
            props.extensionsController,
            getInsightCombinedViews,
        ])
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

    const configureURL = isCreationUIEnabled ? '/insights/create' : '/user/settings'

    // Remove uniqBy when this extension api issue will be resolved
    // https://github.com/sourcegraph/sourcegraph/issues/20442
    const filteredViews = useMemo(() => {
        if (!views) {
            return views
        }

        return uniqBy(views, view => view.id)
    }, [views])

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
                            <Link to={configureURL} onClick={logConfigureClick} className="btn btn-secondary">
                                <GearIcon className="icon-inline" /> Configure insights
                            </Link>
                        </>
                    }
                    className="mb-3"
                />
                {filteredViews === undefined ? (
                    <div className="d-flex w-100">
                        <LoadingSpinner className="my-4" />
                    </div>
                ) : (
                    <InsightsViewGrid {...props} views={filteredViews} />
                )}
            </Page>
        </div>
    )
}
