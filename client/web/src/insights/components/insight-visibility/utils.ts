import { RealInsightDashboard } from '../../core/types';

/**
 * Create dashboard map with dashboard id as a key and dashboard itself as a value
 *
 * @param dashboards - list of dashboards
 */
export const createDashboardsMap = (...dashboards: RealInsightDashboard[]) : Record<string, RealInsightDashboard> =>
    dashboards.reduce<Record<string, RealInsightDashboard>>(
        (dashboards, dashboard) => ({...dashboards, [dashboard.id]: dashboard}),
        {}
    )
