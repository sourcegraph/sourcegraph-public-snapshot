import * as uuid from 'uuid'

import { InsightDashboard } from '../../../../../../schema/settings.schema'
import { DashboardCreationFields } from '../components/insights-dashboard-creation-content/InsightsDashboardCreationContent'

/**
 * Creates sanitized dashboard configuration object according to public user setting's API.
 *
 * @param values - a dashboard creation UI form values
 */
export function createSanitizedDashboard(values: DashboardCreationFields): InsightDashboard {
    return {
        id: uuid.v4(),
        title: values.name.trim(),
        insightIds: [],
    }
}
