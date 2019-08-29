import React from 'react'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { CAMPAIGN_TEMPLATES, CampaignTemplate } from '../../form/templates'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface Props extends CampaignFormProps {}

/**
 * A form to edit a campaign.
 */
export const EditCampaignForm: React.FunctionComponent<Props> = props => {
    const templateID = props.value.template ? props.value.template.template : null
    const template: null | CampaignTemplate | ErrorLike =
        templateID !== null
            ? CAMPAIGN_TEMPLATES.find(({ id }) => id === templateID) ||
              new Error('Template not found. Please choose a template from the list.')
            : null
    const TemplateIcon = template !== null && !isErrorLike(template) ? template.icon : undefined
    const TemplateForm = template !== null && !isErrorLike(template) ? template.renderForm : undefined

    return (
        <CampaignForm {...props}>
            {({ form }) => (
                <>
                    <h2>Edit campaign</h2>
                    {form}
                    {template === null ? (
                        <div className="alert alert-danger">Invalid campaign template</div>
                    ) : isErrorLike(template) ? (
                        <div className="alert alert-danger">{template.message}</div>
                    ) : (
                        !template.isEmpty && (
                            <>
                                <h4 className="d-flex align-items-start">
                                    {TemplateIcon && <TemplateIcon className="icon-inline mr-2 flex-0" />} Edit:{' '}
                                    {template.title}
                                </h4>
                                <p>{template.detail}</p>
                                {TemplateForm && <TemplateForm {...props} />}
                            </>
                        )
                    )}
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
}
