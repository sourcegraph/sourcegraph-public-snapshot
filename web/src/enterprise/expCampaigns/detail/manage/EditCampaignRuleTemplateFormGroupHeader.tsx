import React, { useCallback, MouseEventHandler } from 'react'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { RuleTemplate } from '../../form/templates'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'
import { CampaignFormControl } from '../../form/CampaignForm'

interface Props extends CampaignFormControl {
    ruleIndex: number
    template: null | RuleTemplate | ErrorLike
}

export const EditCampaignRuleTemplateFormGroupHeader: React.FunctionComponent<Props> = ({
    template,
    ruleIndex,
    value: campaignValue,
    onChange: onCampaignChange,
    disabled,
}) => {
    const TemplateIcon = template !== null && !isErrorLike(template) ? template.icon : undefined

    const onRemoveClick = useCallback<MouseEventHandler<HTMLButtonElement>>(
        e => {
            e.preventDefault()
            if (!campaignValue.rules) {
                return // unreachable
            }
            onCampaignChange({
                rules: [...campaignValue.rules.slice(0, ruleIndex), ...campaignValue.rules.slice(ruleIndex + 1)],
            })
        },
        [campaignValue.rules, onCampaignChange, ruleIndex]
    )

    return template === null ? (
        <div className="alert alert-danger">Invalid campaign template</div>
    ) : isErrorLike(template) ? (
        <div className="alert alert-danger">{template.message}</div>
    ) : !template.isEmpty ? (
        <>
            <header className="d-flex align-items-center justify-content-between mb-2">
                <h3 className="mb-0 d-flex align-items-start">
                    {TemplateIcon && <TemplateIcon className="icon-inline mr-2 flex-0" />} Edit: {template.title}
                </h3>
                {/* TODO!(sqs): show Remove button in new campaign form too */}
                <button type="button" className="btn btn-sm btn-secondary" onClick={onRemoveClick} disabled={disabled}>
                    Remove
                </button>
            </header>
            <p>{template.detail && <Markdown dangerousInnerHTML={renderMarkdown(template.detail)} inline={true} />}</p>
        </>
    ) : null
}
