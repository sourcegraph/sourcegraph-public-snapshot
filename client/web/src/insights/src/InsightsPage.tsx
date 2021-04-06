import React, { useCallback, useEffect } from 'react'
import { ExtensionsControllerProps } from '@sourcegraph/shared/out/src/extensions/controller'
import { InsightsViewGrid, InsightsViewGridProps } from './components/InsightsViewGrid/InsightsViewGrid'
import { InsightsIcon } from './components'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageHeader } from '../../components/PageHeader'
import { FeedbackBadge } from '../../components/FeedbackBadge'
import { Page } from '../../components/Page'
import { TelemetryProps } from '@sourcegraph/shared/out/src/telemetry/telemetryService'
import { ViewInsightProviderResult } from './core/backend'
import { Link } from 'react-router-dom';
import PlusIcon from 'mdi-react/PlusIcon';
import GearIcon from 'mdi-react/GearIcon';

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


