import React, { useCallback } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignFormControl } from '../CampaignForm'
import { RuleTemplate, RULE_TEMPLATES } from '.'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'

interface Props extends CampaignFormControl {
    ruleIndex: number

    url: string
    header?: React.FunctionComponent<
        { url: string; template: null | RuleTemplate | ErrorLike; ruleIndex: number } & CampaignFormControl
    >
}

/**
 * The form group for editing the template that defines a rule.
 */
export const RuleTemplateFormGroup: React.FunctionComponent<Props> = ({
    value,
    onChange,
    ruleIndex,
    url,
    header,
    ...props
}) => {
    const rule = (value.rules && value.rules[ruleIndex]) || null
    const ruleTemplate = rule && rule.template ? rule.template.template : null

    const template: null | RuleTemplate | ErrorLike = ruleTemplate
        ? RULE_TEMPLATES.find(({ id }) => id === ruleTemplate) ||
          new Error('Template not found. Please choose a template from the list.')
        : null

    const TemplateForm = template !== null && !isErrorLike(template) ? template.renderForm : undefined

    const onRuleChange = useCallback(
        (rule: GQL.INewRuleInput): void => {
            onChange({
                rules: [...(value.rules || []).slice(0, ruleIndex), rule, ...(value.rules || []).slice(ruleIndex + 1)],
            })
        },
        [onChange, ruleIndex, value.rules]
    )

    return (
        <>
            {header && header({ url, template, value, onChange, ruleIndex, ...props })}
            {rule && TemplateForm ? (
                <TemplateForm value={rule} onChange={onRuleChange} onCampaignChange={onChange} location={location} />
            ) : (
                <div className="alert alert-danger">Unable to show form for template.</div>
            )}
        </>
    )
}
