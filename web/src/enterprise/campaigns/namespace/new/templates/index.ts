import { PackageJsonDeprecationCampaignTemplate } from './PackageJsonDeprecationCampaignTemplate'

export interface CampaignTemplateContext {}

export interface CampaignTemplate {
    id: string
    title: string
    detail: string
    icon: React.ComponentType<{ className?: string }>
    renderForm: React.FunctionComponent<CampaignTemplateContext>
}

export const CAMPAIGN_TEMPLATES: CampaignTemplate[] = [PackageJsonDeprecationCampaignTemplate]
