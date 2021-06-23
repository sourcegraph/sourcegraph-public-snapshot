import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { PageHeader } from '@sourcegraph/wildcard/src'

import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Page } from '../../../components/Page'
import { CodeInsightsIcon } from '../../components'

export interface DashboardsPageProps {}

/**
 * Displays insights dashboard page - dashboard selector and grid of insights from the dashboard.
 */
export const DashboardsPage: React.FunctionComponent<DashboardsPageProps> = () => (
    <div className="w-100">
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
            Dashboard content
        </Page>
    </div>
)
