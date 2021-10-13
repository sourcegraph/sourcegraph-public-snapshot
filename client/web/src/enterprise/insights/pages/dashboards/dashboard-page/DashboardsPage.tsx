import PlusIcon from 'mdi-react/PlusIcon'
import React, { useEffect, useRef, useState } from 'react'
import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, PageHeader } from '@sourcegraph/wildcard'

import { Badge } from '../../../../../components/Badge'
import { Page } from '../../../../../components/Page'
import { FeedbackPromptContent } from '../../../../../nav/Feedback/FeedbackPrompt'
import { CodeInsightsIcon } from '../../../components'
import { flipRightPosition } from '../../../components/context-menu/utils'
import { Popover } from '../../../components/popover/Popover'
import { InsightsDashboardType } from '../../../core/types'

import { DashboardsContent } from './components/dashboards-content/DashboardsContent'
import styles from './DashboardPage.module.scss'

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
        return <Redirect to={`${url}/${InsightsDashboardType.All}`} />
    }

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<PageAnnotation />}
                    path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
                    actions={
                        <>
                            <Link to="/insights/add-dashboard" className="btn btn-outline-secondary mr-2">
                                <PlusIcon className="icon-inline" /> Create new dashboard
                            </Link>
                            <Link
                                to={`/insights/create?dashboardId=${dashboardID}`}
                                className="btn btn-secondary"
                                onClick={handleAddMoreInsightClick}
                            >
                                <PlusIcon className="icon-inline" /> Create new insight
                            </Link>
                        </>
                    }
                    className="mb-3"
                />

                <DashboardsContent telemetryService={telemetryService} dashboardID={dashboardID} />
            </Page>
        </div>
    )
}

const PageAnnotation: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center">
            <a href="https://docs.sourcegraph.com/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <Badge status="beta" className="text-uppercase" />
            </a>

            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>

            <Popover
                isOpen={isVisible}
                target={buttonReference}
                position={flipRightPosition}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Code Insights: "
                    routeMatch="/insights/dashboards"
                />
            </Popover>
        </div>
    )
}
