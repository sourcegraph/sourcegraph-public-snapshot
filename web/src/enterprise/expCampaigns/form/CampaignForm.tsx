import React, { useCallback, useEffect } from 'react'
import H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'
import { JSONSchema7 } from 'json-schema'
import { Workflow } from '../../../schema/workflow.schema'
import { RuleTemplate } from './templates'
import { parseJSONCOrError } from '../../../../../shared/src/util/jsonc'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { WorkflowEditor } from './WorkflowEditor'
import { ThemeProps } from '../../../theme'

export interface CampaignFormData extends Omit<GQL.IExpCreateCampaignInput, 'name' | 'extensionData' | 'rules'> {
    name: string | null
    nameSuggestion?: string
    draft: boolean
    workflowAsJSONCString: string
    startDate?: string // TODO!(sqs): implement
    dueDate?: string // TODO!(sqs): implement
}

export interface CampaignFormControl {
    value: CampaignFormData
    workflowJSONSchema?: JSONSchema7
    isValid: boolean
    onChange: (value: Partial<CampaignFormData>) => void
    disabled?: boolean
    isLoading?: boolean
}

export interface CampaignFormProps extends CampaignFormControl, ThemeProps {
    /** Called when the form is submitted. */
    onSubmit: () => void

    template?: RuleTemplate

    className?: string
    history: H.History
}

/**
 * A form to create or edit a campaign.
 */
export const CampaignForm: React.FunctionComponent<
    CampaignFormProps & {
        children: ({ form, template }: { form: React.ReactFragment; template: RuleTemplate | undefined }) => JSX.Element
    }
> = ({
    value,
    isValid,
    onChange,
    onSubmit: parentOnSubmit,
    disabled,
    isLoading,
    template,
    className = '',
    children,
    ...props
}) => {
    const onSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            parentOnSubmit()
        },
        [parentOnSubmit]
    )

    useEffect(() => {
        if (template && template.suggestTitle) {
            const workflow = parseJSONCOrError<Workflow>(value.workflowAsJSONCString)
            if (!isErrorLike(workflow)) {
                const suggestedTitle = template.suggestTitle(workflow)
                if (suggestedTitle !== undefined) {
                    const update: Partial<CampaignFormData> = {}
                    if (suggestedTitle !== value.nameSuggestion) {
                        update.nameSuggestion = suggestedTitle
                    }
                    if (suggestedTitle === value.name) {
                        update.name = null
                    }
                    if (Object.keys(update).length > 0) {
                        onChange(update)
                    }
                }
            }
        }
    }, [onChange, template, value.name, value.nameSuggestion, value.workflowAsJSONCString])

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <style>{'.form-group { max-width: 45rem; }' /* TODO!(sqs): hack */}</style>
            {children({
                form: (
                    <>
                        <CampaignFormCommonFields
                            value={value}
                            isValid={isValid}
                            onChange={onChange}
                            disabled={disabled}
                            isLoading={isLoading}
                            autoFocus={true}
                            className="mt-4"
                        />
                        {template && !template.noWorkflow && (
                            <WorkflowEditor
                                {...props}
                                value={value}
                                isValid={isValid}
                                onChange={onChange}
                                disabled={disabled}
                                isLoading={isLoading}
                            />
                        )}
                    </>
                ),
                template,
            })}
        </Form>
    )
}
