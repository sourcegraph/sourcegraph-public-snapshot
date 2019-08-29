import H from 'history'
import React from 'react'
import { CampaignFormControl } from '../CampaignForm'
import { CodeOwnershipValidationCampaignTemplate } from './CodeOwnershipValidationCampaignTemplate'
import { ESLintRuleCampaignTemplate } from './ESLintRuleCampaignTemplate'
import { ExistingExternalChangesetsAndIssuesCampaignTemplate } from './ExistingExternalChangesetsAndIssuesCampaignTemplate'
import { FindReplaceCampaignTemplate } from './FindReplaceCampaignTemplate'
import { JavaArtifactDependencyCampaignTemplate } from './JavaArtifactDependencyCampaignTemplate'
import { PackageJsonDependencyCampaignTemplate } from './PackageJsonDependencyCampaignTemplate'
import { TriageSearchResultsCampaignTemplate } from './TriageSearchResultsCampaignTemplate'

export interface CampaignTemplateComponentContext extends CampaignFormControl {
    location?: Pick<H.Location, 'search'>
}

export interface CampaignTemplate {
    id: string
    title: string
    detail?: string
    icon?: React.ComponentType<{ className?: string }>
    renderForm: React.FunctionComponent<CampaignTemplateComponentContext>
    isEmpty?: boolean
}

export const EMPTY_CAMPAIGN_TEMPLATE_ID = 'empty'

export const CAMPAIGN_TEMPLATES: CampaignTemplate[] = [
    JavaArtifactDependencyCampaignTemplate,
    PackageJsonDependencyCampaignTemplate,
    ESLintRuleCampaignTemplate,
    CodeOwnershipValidationCampaignTemplate,
    FindReplaceCampaignTemplate,
    TriageSearchResultsCampaignTemplate,
    ExistingExternalChangesetsAndIssuesCampaignTemplate,
    { id: EMPTY_CAMPAIGN_TEMPLATE_ID, title: '', renderForm: () => null, isEmpty: true },
]
