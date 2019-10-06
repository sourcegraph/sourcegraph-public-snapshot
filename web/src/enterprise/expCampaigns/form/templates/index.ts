import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import { CodeOwnershipValidationRuleTemplate } from './CodeOwnershipValidationRuleTemplate'
import { ESLintRuleTemplate } from './ESLintRuleTemplate'
import { ExistingExternalChangesetsAndIssuesRuleTemplate } from './ExistingExternalChangesetsAndIssuesRuleTemplate'
import { FindReplaceRuleTemplate } from './FindReplaceRuleTemplate'
import { JavaArtifactDependencyRuleTemplate } from './JavaArtifactDependencyRuleTemplate'
import { PackageJsonDependencyRuleTemplate } from './PackageJsonDependencyRuleTemplate'
import { TriageSearchResultsRuleTemplate } from './TriageSearchResultsRuleTemplate'
import { CampaignFormData } from '../CampaignForm'
import { USE_CAMPAIGN_RULES } from '../..'
import { RubyGemDependencyRuleTemplate } from './RubyGemDependencyRuleTemplate'

export interface RuleTemplateComponentContext {
    value: GQL.INewRuleInput
    onChange: (value: GQL.INewRuleInput) => void
    onCampaignChange: (value: Partial<CampaignFormData>) => void

    disabled?: boolean
    isLoading?: boolean

    location?: Pick<H.Location, 'search'>
}

export interface RuleTemplate {
    id: string
    title: string
    detail?: string
    icon?: React.ComponentType<{ className?: string }>
    renderForm: React.FunctionComponent<RuleTemplateComponentContext>
    isEmpty?: boolean
}

export const EMPTY_RULE_TEMPLATE_ID = 'empty'

export const RULE_TEMPLATES: RuleTemplate[] = [
    ...(USE_CAMPAIGN_RULES
        ? [
              JavaArtifactDependencyRuleTemplate,
              PackageJsonDependencyRuleTemplate,
              RubyGemDependencyRuleTemplate,
              ESLintRuleTemplate,
              CodeOwnershipValidationRuleTemplate,
              FindReplaceRuleTemplate,
              TriageSearchResultsRuleTemplate,
          ]
        : []),
    ExistingExternalChangesetsAndIssuesRuleTemplate,
    { id: EMPTY_RULE_TEMPLATE_ID, title: '', renderForm: () => null, isEmpty: true },
]
