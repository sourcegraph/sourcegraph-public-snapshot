import H from 'history'
import React from 'react'
import { USE_CAMPAIGN_RULES } from '../..'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignFormData } from '../CampaignForm'
import { ExistingExternalChangesetsAndIssuesRuleTemplate } from './ExistingExternalChangesetsAndIssuesRuleTemplate'
import { FindReplaceRuleTemplate } from './FindReplaceRuleTemplate'
import { NPMCredentialsRuleTemplate } from './NPMCredentialsRuleTemplate'
import { PackageJsonDependencyRuleTemplate } from './PackageJsonDependencyRuleTemplate'
import { RubyGemDependencyRuleTemplate } from './RubyGemDependencyRuleTemplate'
import { TriageSearchResultsRuleTemplate } from './TriageSearchResultsRuleTemplate'
import { JavaDependencyRuleTemplate } from './JavaDependencyRuleTemplate'
import { Workflow } from '../../../../schema/workflow.schema'
import { JSONSchema7 } from 'json-schema'

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
    defaultWorkflow?: Workflow
    workflowJSONSchema?: JSONSchema7
    suggestTitle?: (workflow: Workflow) => string | undefined
}

export const EMPTY_RULE_TEMPLATE_ID = 'empty'

export const RULE_TEMPLATES: RuleTemplate[] = [
    ...(USE_CAMPAIGN_RULES
        ? [
              PackageJsonDependencyRuleTemplate,
              JavaDependencyRuleTemplate,
              NPMCredentialsRuleTemplate,
              // RubyGemDependencyRuleTemplate,
              FindReplaceRuleTemplate,
              // TriageSearchResultsRuleTemplate,
          ]
        : []),
    ExistingExternalChangesetsAndIssuesRuleTemplate,
    { id: EMPTY_RULE_TEMPLATE_ID, title: '', renderForm: () => null, isEmpty: true },
]
