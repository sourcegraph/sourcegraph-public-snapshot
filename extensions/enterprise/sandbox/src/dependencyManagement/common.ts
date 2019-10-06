export const LOADING = 'loading' as const

export interface DependencyManagementCampaignContextCommon {
    packageName?: string
    matchVersion?: string
    action: 'ban' | { requireVersion: string }
    createChangesets: boolean
    filters?: string
}
