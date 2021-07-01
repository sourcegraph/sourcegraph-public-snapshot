import PlusIcon from 'mdi-react/PlusIcon'
import React, { useContext, useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard/src'

import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Page } from '../../../components/Page'
import { CodeInsightsIcon, InsightsViewGrid, InsightsViewGridProps } from '../../components'
import { InsightsApiContext } from '../../core/backend/api-provider'

export interface DashboardsPageProps
    extends Omit<InsightsViewGridProps, 'views'>,
        SettingsCascadeProps,
        ExtensionsControllerProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from user/org by id. In case if id equals undefined
     * we will get insights from final version of merged settings (all insights)
     */
    dashboardID?: string
}

/**
 * Displays insights dashboard page - dashboard selector and grid of insights from the dashboard.
 */
export const DashboardsPage: React.FunctionComponent<DashboardsPageProps> = props => {
    const { dashboardID, settingsCascade, extensionsController } = props
    const { getInsightCombinedViews } = useContext(InsightsApiContext)

    const insightIds = useMemo(() => {
        if (isErrorLike(settingsCascade.final) || !settingsCascade.final || !dashboardID) {
            return undefined
        }

        const dashboardConfiguration = settingsCascade.final[`insightDashboard.${dashboardID}`]

        if (!dashboardConfiguration) {
            return []
        }

        return dashboardConfiguration.ids
    }, [dashboardID, settingsCascade])

    const views = useObservable(
        useMemo(() => getInsightCombinedViews(extensionsController?.extHostAPI, insightIds), [
            extensionsController,
            insightIds,
            getInsightCombinedViews,
        ])
    )

    return (
        <Page>
            <PageHeader
                annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
                actions={
                    <Link to="/insights/create" className="btn btn-secondary mr-1">
                        <PlusIcon className="icon-inline" /> Create new insight
                    </Link>
                }
                className="mb-3"
            />
            {views === undefined ? (
                <div className="d-flex w-100">
                    <LoadingSpinner className="my-4" />
                </div>
            ) : (
                <InsightsViewGrid {...props} views={views} hasContextMenu={true} />
            )}
        </Page>
    )
}
