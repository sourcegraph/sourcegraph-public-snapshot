import EslintIcon from 'mdi-react/EslintIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'
import { RuleDefinition } from '../../../../rules/form/definition/RuleDefinitionFormControl'

interface Props extends CampaignTemplateComponentContext {}

const ESLintRuleCampaignTemplateForm: React.FunctionComponent<Props> = ({ onChange, disabled, location }) => {
    const params = new URLSearchParams(location.search)
    const [rules, setRules] = useState(params.get('rules') || '')
    const onRulesChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setRules(e.currentTarget.value)
    }, [])

    useEffect(() => {
        const rulesOrPlaceholder = rules || '<rule>'
        onChange({
            name: `Enforce ESLint rule${/[\s,]/.test(rulesOrPlaceholder) ? 's:' : ''} ${rulesOrPlaceholder}`,
            rules: [
                // TODO!(sqs): hack
                {
                    name: 'Warn on ESLint rule violations',
                    // tslint:disable-next-line: no-object-literal-type-assertion
                    definition: JSON.stringify({
                        conditions: `is:diagnostic diagnostic.type:eslint eslint.rule:${rulesOrPlaceholder.replace(
                            /\s/g,
                            ''
                        )}`,
                    } as RuleDefinition),
                },
            ],
        })
    }, [onChange, rules])

    return (
        <>
            <div className="form-group">
                <label htmlFor="campaign-template-form__rules">Rule IDs</label>
                <input
                    type="text"
                    id="campaign-template-form__rules"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="no-console, no-extra-parens"
                    value={rules}
                    onChange={onRulesChange}
                    autoFocus={true}
                    disabled={disabled}
                />
                <small className="form-help text-muted">
                    <a href="https://eslint.org/docs/rules/" target="_blank">
                        List of ESLint builtin rules
                    </a>
                </small>
            </div>
        </>
    )
}

export const ESLintRuleCampaignTemplate: CampaignTemplate = {
    id: 'eslintRule',
    title: 'Gradually enforce new ESLint rule',
    detail:
        'Warn on violations of a new ESLint rule and open changesets to fix all problems and add the rule to .eslintrc files.',
    icon: EslintIcon,
    renderForm: ESLintRuleCampaignTemplateForm,
}
