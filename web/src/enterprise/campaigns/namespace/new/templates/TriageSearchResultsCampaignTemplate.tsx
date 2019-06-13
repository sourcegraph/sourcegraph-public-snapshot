import FileSearchIcon from 'mdi-react/FileSearchIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const TriageSearchResultsCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const TriageSearchResultsCampaignTemplate: CampaignTemplate = {
    id: 'triageSearchResults',
    title: 'Triage from search results',
    detail: 'Collect search results for a query, assigning and manually reviewing each one.',
    icon: FileSearchIcon,
    renderForm: TriageSearchResultsCampaignTemplateForm,
}
