import PlusIcon from 'mdi-react/PlusIcon'
import React, { useEffect } from 'react'
import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button } from '@sourcegraph/wildcard'

import { Page } from '../../../../../components/Page'
import { CodeInsightsIcon } from '../../../components'
import { BetaFeedbackPanel } from '../../../components/beta-feedback-panel/BetaFeedbackPanel'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../../core/types/dashboard/virtual-dashboard'

import { DashboardsContent } from './components/dashboards-content/DashboardsContent'

export interface DashboardsPageProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID?: string
}

/**
 * Displays insights dashboard page - dashboard selector and grid of dashboard insights.
 */
export const DashboardsPage: React.FunctionComponent<DashboardsPageProps> = props => {
    const { dashboardID, telemetryService } = props
    const { url } = useRouteMatch()

    useEffect(() => {
        telemetryService.logViewEvent('Insights')
    }, [telemetryService, dashboardID])

    const handleAddMoreInsightClick = (): void => {
        telemetryService.log('InsightAddMoreClick')
    }

    if (!dashboardID) {
        // In case if url doesn't have a dashboard id we should fallback on
        // built-in "All insights" dashboard
        return <Redirect to={`${url}/${ALL_INSIGHTS_DASHBOARD_ID}`} />
    }

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<BetaFeedbackPanel />}
                    path={[{ icon: CodeInsightsIcon }, { text: 'Insights' }]}
                    actions={
                        <>
                            <Button
                                to="/insights/add-dashboard"
                                className="mr-2"
                                variant="secondary"
                                outline={true}
                                as={Link}
                            >
                                <PlusIcon className="icon-inline" /> Create new dashboard
                            </Button>
                            <Button
                                to={`/insights/create?dashboardId=${dashboardID}`}
                                onClick={handleAddMoreInsightClick}
                                variant="secondary"
                                as={Link}
                            >
                                <PlusIcon className="icon-inline" /> Create new insight
                            </Button>
                        </>
                    }
                    className="mb-3"
                />

                <DashboardsContent telemetryService={telemetryService} dashboardID={dashboardID} />
            </Page>
        </div>
    )
}
