import React, { useCallback, useEffect, useMemo } from 'react'
import { useObservable } from '@sourcegraph/shared/out/src/util/useObservable'
import { ExtensionsControllerProps } from '@sourcegraph/shared/out/src/extensions/controller'
import { InsightsViewGrid, InsightsViewGridProps } from './components/InsightsViewGrid/InsightsViewGrid'
import { InsightsIcon } from './components'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from '@sourcegraph/shared/out/src/components/Link'
import GearIcon from 'mdi-react/GearIcon'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageHeader } from '../../components/PageHeader'
import { FeedbackBadge } from '../../components/FeedbackBadge'
import { Page } from '../../components/Page'
import { TelemetryProps } from '@sourcegraph/shared/out/src/telemetry/telemetryService'
import {getCombinedViews, ViewInsightProviderResult} from './core/backend'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { wrapRemoteObservable } from '@sourcegraph/shared/out/src/api/client/api/common'

interface InsightsPageProps extends ExtensionsControllerProps, Omit<InsightsViewGridProps, 'views'>, TelemetryProps {
    views: ViewInsightProviderResult[]
}

// Main entry point to code insight page
export const InsightPage: React.FunctionComponent<InsightsPageProps> = props => {
    const views = props.views ?? [];

    const logConfigureClick = useCallback(() => {
        props.telemetryService.log('InsightConfigureClick')
    }, [props.telemetryService])

    const logAddMoreClick = useCallback(() => {
        props.telemetryService.log('InsightAddMoreClick')
    }, [props.telemetryService])

    useEffect(() => {
        props.telemetryService.logViewEvent('Insights')
    }, [props.telemetryService])

    return (
        <InsightsPageContent
            {...props}
            logConfigureClick={logConfigureClick}
            logAddMoreClick={logAddMoreClick}
            views={views}/>
    );
}

interface InsightsPageContentProps extends InsightsPageProps {
    views?: ViewInsightProviderResult[];
    logConfigureClick?: () => void;
    logAddMoreClick?: () => void;
}

export const InsightsPageContent: React.FunctionComponent<InsightsPageContentProps> = props => {

    const { views, logAddMoreClick, logConfigureClick, } = props;

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                    path={[{ icon: InsightsIcon, text: 'Code insights' }]}
                    actions={null}
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


