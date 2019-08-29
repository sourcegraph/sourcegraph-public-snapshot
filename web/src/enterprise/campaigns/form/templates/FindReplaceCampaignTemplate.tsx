import FindReplaceIcon from 'mdi-react/FindReplaceIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const FindReplaceCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const FindReplaceCampaignTemplate: CampaignTemplate = {
    id: 'findReplace',
    title: 'Find-replace',
    detail: 'Configurable find-replace across multiple files and repositories, opening changesets with the diffs.',
    icon: FindReplaceIcon,
    renderForm: FindReplaceCampaignTemplateForm,
}
