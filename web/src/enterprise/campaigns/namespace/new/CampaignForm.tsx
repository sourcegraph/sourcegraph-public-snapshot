import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { Form } from '../../../../components/Form'
import { useLocalStorage } from '../../../../util/useLocalStorage'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'
import { CampaignTemplateChooser } from './CampaignTemplateChooser'
import { CAMPAIGN_TEMPLATES, CampaignTemplate, EMPTY_CAMPAIGN_TEMPLATE_ID } from './templates'

export interface CampaignFormData
    extends Pick<GQL.ICreateCampaignInput, Exclude<keyof GQL.ICreateCampaignInput, 'extensionData'>> {
    isValid: boolean
    dueDate?: string
}

export interface CampaignFormControl {
    value: CampaignFormData
    onChange: (value: Partial<CampaignFormData>) => void
    disabled: boolean
}

interface Props {
    templateID: string | null

    initialValue: CampaignFormData

    /** Called when the form's data is changed. */
    onChange: (data: CampaignFormData) => void

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (data: CampaignFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string

    /** The URL of the form. */
    match: {
        url: string
    }
    location: H.Location
}

/**
 * A form to create or edit a campaign.
 */
export const CampaignForm: React.FunctionComponent<Props> = ({
    templateID,
    initialValue,
    onChange: parentOnChange,
    onDismiss,
    onSubmit: parentOnSubmit,
    buttonText,
    isLoading,
    className = '',
    match,
    location,
}) => {
    const [value, setValue] = useState<CampaignFormData>(initialValue)
    const onChange = useCallback((value: Partial<CampaignFormData>) => {
        setValue(prevValue => ({ ...prevValue, ...value }))
    }, [])
    useEffect(() => parentOnChange(value), [parentOnChange, value])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            parentOnSubmit(value)
        },
        [parentOnSubmit, value]
    )

    const template: null | CampaignTemplate | ErrorLike =
        templateID !== null
            ? CAMPAIGN_TEMPLATES.find(({ id }) => id === templateID) ||
              new Error('Template not found. Please choose a template from the list.')
            : null
    const formControlProps: CampaignFormControl = {
        value,
        onChange,
        disabled: isLoading,
    }

    const urlToFormWithTemplate = useCallback(
        (templateID: string) => `${match.url}?${new URLSearchParams({ template: templateID }).toString()}`,
        [match.url]
    )
    const TemplateIcon = template !== null && !isErrorLike(template) ? template.icon : undefined
    const TemplateForm = template !== null && !isErrorLike(template) ? template.renderForm : undefined

    const [isCreateCampaignInputVisible, setIsCreateCampaignInputVisible] = useLocalStorage(
        'CampaignForm.isCreateCampaignInputVisible',
        false
    )
    const toggleIsCreateCampaignInputVisible = useCallback(
        () => setIsCreateCampaignInputVisible(!isCreateCampaignInputVisible),
        [isCreateCampaignInputVisible, setIsCreateCampaignInputVisible]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <style>{`.form-group { max-width: 45rem; }` /* TODO!(sqs): hack */}</style>
            {template === null || isErrorLike(template) ? (
                <>
                    {isErrorLike(template) && <div className="alert alert-danger">{template.message}</div>}
                    <CampaignTemplateChooser
                        {...formControlProps}
                        urlToFormWithTemplate={urlToFormWithTemplate}
                        className="mb-4"
                    />
                    <p>
                        Don't see what you're looking for?{' '}
                        <Link to={urlToFormWithTemplate(EMPTY_CAMPAIGN_TEMPLATE_ID)}>Create a new empty campaign</Link>{' '}
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
                                {TemplateIcon && <TemplateIcon className="icon-inline mr-2 flex-0" />} New campaign:{' '}
                                {template.title}
                            </h2>
                            <p>
                                {template.detail}{' '}
                                <Link to={match.url} className="text-muted mb-2">
                                    Choose a different template.
                                </Link>
                            </p>
                            {TemplateForm && <TemplateForm {...formControlProps} location={location} />}
                        </>
                    )}
                    <CampaignFormCommonFields {...formControlProps} className="mt-4" />
                    <div className="form-group mt-4" style={{ minHeight: '2.5rem' }}>
                        {onDismiss && (
                            <button
                                type="reset"
                                className="btn btn-secondary mr-2"
                                onClick={onDismiss}
                                disabled={formControlProps.disabled}
                            >
                                Cancel
                            </button>
                        )}
                        <button
                            type="submit"
                            className="btn btn-success"
                            disabled={!value.isValid || formControlProps.disabled}
                        >
                            {isLoading ? <LoadingSpinner className="icon-inline" /> : buttonText}
                        </button>
                        <button
                            type="button"
                            className="btn btn-sm btn-link ml-2"
                            onClick={toggleIsCreateCampaignInputVisible}
                        >
                            {isCreateCampaignInputVisible ? 'Hide' : 'Show'} JSON
                        </button>
                    </div>
                    {isCreateCampaignInputVisible && (
                        <pre className="small mt-4 border p-2 overflow-auto">
                            <code>{JSON.stringify(value, null, 2)}</code>
                        </pre>
                    )}
                </>
            )}
        </Form>
    )
}
