import H from 'history'
import { CampaignFormControl } from '../CampaignForm'
import { CodeOwnershipValidationCampaignTemplate } from './CodeOwnershipValidationCampaignTemplate'
import { ESLintRuleCampaignTemplate } from './ESLintRuleCampaignTemplate'
import { FindReplaceCampaignTemplate } from './FindReplaceCampaignTemplate'
import { JavaArtifactDependencyCampaignTemplate } from './JavaArtifactDependencyCampaignTemplate'
import { PackageJsonDependencyCampaignTemplate } from './PackageJsonDependencyCampaignTemplate'
import { TriageSearchResultsCampaignTemplate } from './TriageSearchResultsCampaignTemplate'

export interface CampaignTemplateComponentContext extends CampaignFormControl {
    location: H.Location
}

export interface CampaignTemplate {
    id: string
    title: string
    detail: string
    icon: React.ComponentType<{ className?: string }>
    renderForm: React.FunctionComponent<CampaignTemplateComponentContext>
}

export const CAMPAIGN_TEMPLATES: CampaignTemplate[] = [
    JavaArtifactDependencyCampaignTemplate,
    PackageJsonDependencyCampaignTemplate,
    ESLintRuleCampaignTemplate,
    CodeOwnershipValidationCampaignTemplate,
    FindReplaceCampaignTemplate,
    TriageSearchResultsCampaignTemplate,
]
