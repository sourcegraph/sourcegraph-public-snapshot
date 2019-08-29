import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { CampaignTemplateChooser } from '../../form/CampaignTemplateChooser'
import { CAMPAIGN_TEMPLATES, CampaignTemplate, EMPTY_CAMPAIGN_TEMPLATE_ID } from '../../form/templates'
import { CreateCampaignButton } from './CreateCampaignButton'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'

interface Props extends CampaignFormProps {
    templateID: string | null

    /** The URL of the form. */
    match: {
        url: string
    }
    location: H.Location
}

/**
 * A form to create a new campaign.
 */
export const NewCampaignForm: React.FunctionComponent<Props> = ({ templateID, match, location, ...props }) => {
    const template: null | CampaignTemplate | ErrorLike =
        templateID !== null
            ? CAMPAIGN_TEMPLATES.find(({ id }) => id === templateID) ||
              new Error('Template not found. Please choose a template from the list.')
            : null

    const urlToFormWithTemplate = useCallback(
        (templateID: string) => `${match.url}?${new URLSearchParams({ template: templateID }).toString()}`,
        [match.url]
    )
    const TemplateIcon = template !== null && !isErrorLike(template) ? template.icon : undefined
    const TemplateForm = template !== null && !isErrorLike(template) ? template.renderForm : undefined

    return (
        <CampaignForm {...props}>
            {({ form }) => (
                <>
                    {template === null || isErrorLike(template) ? (
                        <>
                            {isErrorLike(template) && <div className="alert alert-danger">{template.message}</div>}
                            <CampaignTemplateChooser
                                {...props}
                                urlToFormWithTemplate={urlToFormWithTemplate}
                                className="mb-4"
                            />
                            <p>
                                Don't see what you're looking for?{' '}
                                <Link to={urlToFormWithTemplate(EMPTY_CAMPAIGN_TEMPLATE_ID)}>
                                    Create a new empty campaign
                                </Link>{' '}
                                and manually add changesets, issues, and rules.
                            </p>
                        </>
                    ) : (
                        <>
                            {template.isEmpty ? (
                                <h2>New campaign</h2>
                            ) : (
                                <>
                                    <h2 className="d-flex align-items-start">
                                        {TemplateIcon && <TemplateIcon className="icon-inline mr-2 flex-0" />} New
                                        campaign: {template.title}
                                    </h2>
                                    <p>
                                        {template.detail}{' '}
                                        <Link to={match.url} className="text-muted mb-2">
                                            Choose a different template.
                                        </Link>
                                    </p>
                                    {TemplateForm && <TemplateForm {...props} location={location} />}
                                </>
                            )}
                            {form}
                            <div className="form-group mt-4">
                                <CreateCampaignButton
                                    {...props}
                                    icon={props.isLoading ? LoadingSpinner : undefined}
                                    className="btn btn-success"
                                    disabled={!props.value.isValid || props.disabled || props.isLoading}
                                />
                            </div>
                        </>
                    )}
                </>
            )}
        </CampaignForm>
    )
}
