import EyePlusIcon from 'mdi-react/EyePlusIcon'
import React from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'

interface Props extends RuleTemplateComponentContext {}

const ExistingExternalChangesetsAndIssuesCampaignTemplateForm: React.FunctionComponent<Props> = () => null

export const ExistingExternalChangesetsAndIssuesRuleTemplate: RuleTemplate = {
    id: 'existingExternalChangesetsAndIssues',
    title: 'Existing pull requests and issues',
    detail: 'Track progress on a group of PRs or issues on GitHub, GitLab, and Bitbucket Server.',
    icon: EyePlusIcon,
    renderForm: ExistingExternalChangesetsAndIssuesCampaignTemplateForm,
}
