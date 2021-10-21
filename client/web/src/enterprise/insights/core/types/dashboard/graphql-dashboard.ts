export interface GraphQlDashboard {
    id: string
    title: string
    views: InsightView[]
}

export interface InsightView {
    id: string
    dataSeries: InsightsSeries[]
}

export interface InsightsSeries {
    label: string
    points: InsightDataPoint[]
    status: InsightsSeriesStatus
}

export interface InsightDataPoint {
    dateTime: string
    value: number
}

export interface InsightsSeriesStatus {
    totalPoints: number
    pendingJobs: number
    completedJobs: number
    failedJobs: number
    backfillQueuedAt: string
}
