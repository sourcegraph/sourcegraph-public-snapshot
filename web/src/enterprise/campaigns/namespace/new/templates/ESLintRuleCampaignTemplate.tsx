import EslintIcon from 'mdi-react/EslintIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const ESLintRuleCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const ESLintRuleCampaignTemplate: CampaignTemplate = {
    id: 'eslintRule',
    title: 'Gradually enforce new ESLint rule',
    detail:
        'Warn on violations of a new ESLint rule and open changesets to fix all problems and add the rule to .eslintrc files.',
    icon: EslintIcon,
    renderForm: ESLintRuleCampaignTemplateForm,
}
