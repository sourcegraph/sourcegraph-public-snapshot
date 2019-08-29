import AccountStarIcon from 'mdi-react/AccountStarIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const CodeOwnershipValidationCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const CodeOwnershipValidationCampaignTemplate: CampaignTemplate = {
    id: 'codeOwnershipValidation',
    title: 'Require valid code owners for all files',
    detail:
        'Warn on files without owners, invalid code owners files, and other problems, opening issues/changesets as needed.',
    icon: AccountStarIcon,
    renderForm: CodeOwnershipValidationCampaignTemplateForm,
}
