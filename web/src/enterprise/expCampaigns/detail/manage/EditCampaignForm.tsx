import React from 'react'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { RuleTemplateFormGroup } from '../../form/templates/RuleTemplateFormGroup'
import { EditCampaignRuleTemplateFormGroupHeader } from './EditCampaignRuleTemplateFormGroupHeader'
import { CampaignFormAddRuleTemplateDropdownButton } from '../../form/CampaignFormAddRuleTemplateDropdownButton'

interface Props extends CampaignFormProps {}

/**
 * A form to edit a campaign.
 */
export const EditCampaignForm: React.FunctionComponent<Props> = props => (
    <CampaignForm {...props}>
        {() => (
            <>
                {props.value.rules &&
                    props.value.rules.map((_rule, i) => (
                        <RuleTemplateFormGroup
                            // eslint-disable-next-line react/no-array-index-key
                            key={i}
                            {...props}
                            ruleIndex={i}
                            url="" // TODO!(sqs): only needed in NewCampaignForm
                            header={EditCampaignRuleTemplateFormGroupHeader}
                        />
                    ))}
                <CampaignFormAddRuleTemplateDropdownButton {...props} />
                <div className="form-group mt-4">
                    <button
                        type="submit"
                        className="btn btn-success"
                        disabled={!props.value.isValid || props.disabled || props.isLoading}
                    >
                        {props.isLoading && <LoadingSpinner className="icon-inline mr-2" />} Save
                    </button>
                </div>
            </>
        )}
    </CampaignForm>
)
