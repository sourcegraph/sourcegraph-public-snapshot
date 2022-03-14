import { getDashboardPermissions } from '../../../pages/dashboards/dashboard-page/utils/get-dashboard-permissions'
import { InsightDashboard } from '../../types'
import { UiFeaturesConfig } from '../code-insights-backend'

import { CodeInsightsGqlBackend } from './code-insights-gql-backend'

export class CodeInsightsGqlBackendLimited extends CodeInsightsGqlBackend {
    public getUiFeatures = (currentDashboard?: InsightDashboard): UiFeaturesConfig => ({
        licensed: false,
        permissions: getDashboardPermissions(currentDashboard, false),
    })
}
