import FileSearchIcon from 'mdi-react/FileSearchIcon'
import React from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'

interface Props extends RuleTemplateComponentContext {}

const TriageSearchResultsCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const TriageSearchResultsRuleTemplate: RuleTemplate = {
    id: 'triageSearchResults',
    title: 'Triage from search results',
    detail: 'Collect search results for a query, assigning and manually reviewing each one.',
    icon: FileSearchIcon,
    renderForm: TriageSearchResultsCampaignTemplateForm,
}
