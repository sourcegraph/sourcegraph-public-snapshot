import EyePlusIcon from 'mdi-react/EyePlusIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const ExistingExternalChangesetsAndIssuesCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => (
    <p>hello world</p>
)

export const ExistingExternalChangesetsAndIssuesCampaignTemplate: CampaignTemplate = {
    id: 'existingExternalChangesetsAndIssues',
    title: 'Existing pull requests and issues',
    detail: 'Track progress on a group of PRs or issues on GitHub, GitLab, and Bitbucket Server.',
    icon: EyePlusIcon,
    renderForm: ExistingExternalChangesetsAndIssuesCampaignTemplateForm,
}
