import { CampaignType } from './backend'

/** The maximum amount of changeset nodes rendered initially */
export const DEFAULT_CHANGESET_LIST_COUNT = 15

export const MANUAL_CAMPAIGN_TYPE = 'manual' as const

export const campaignTypeLabels: Record<CampaignType | typeof MANUAL_CAMPAIGN_TYPE, string> = {
    [MANUAL_CAMPAIGN_TYPE]: 'Manual',
    comby: 'Comby search and replace',
    credentials: 'Find leaked credentials',
}
