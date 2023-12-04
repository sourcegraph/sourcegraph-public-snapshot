export enum CodeInsightsPageRoutes {
    InsightsDashboardTab = '/dashboards/:dashboardId?',
    InsightsAll = '/dashboards/all',
    InsightsAboutTab = '/about',
}

export function encodeDashboardIdQueryParam(url: string, dashboardId?: string): string {
    // Skip encoding if we're dealing with all insights dashboard,
    // or we don't have any dashboard id.
    if (!dashboardId || dashboardId === 'all') {
        return url
    }

    return `${url}?dashboardId=${dashboardId}`
}
