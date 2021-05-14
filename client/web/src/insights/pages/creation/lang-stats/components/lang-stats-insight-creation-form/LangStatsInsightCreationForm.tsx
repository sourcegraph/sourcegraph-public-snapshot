import classnames from 'classnames'
import React, { FormEventHandler, RefObject } from 'react'

import { ErrorAlert } from '../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../components/LoaderButton'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput'
import { useFieldAPI } from '../../../../../components/form/hooks/useField'
import { FORM_ERROR, SubmissionErrors } from '../../../../../components/form/hooks/useForm'
import { LangStatsCreationFormFields } from '../../types'

import styles from './LangStatsInsightCreationForm.module.scss'

export interface LangStatsInsightCreationFormProps {
    mode?: 'creation' | 'edit'
    innerRef: RefObject<any>
    handleSubmit: FormEventHandler
    submitErrors: SubmissionErrors
    submitting: boolean
    className?: string

    title: useFieldAPI<LangStatsCreationFormFields['title']>
    repository: useFieldAPI<LangStatsCreationFormFields['repository']>
    threshold: useFieldAPI<LangStatsCreationFormFields['threshold']>
    visibility: useFieldAPI<LangStatsCreationFormFields['visibility']>

    onCancel: () => void
}

export const LangStatsInsightCreationForm: React.FunctionComponent<LangStatsInsightCreationFormProps> = props => {
    const {
        mode = 'creation',
        innerRef,
        handleSubmit,
        submitErrors,
        submitting,
        className,
        title,
        repository,
        threshold,
        visibility,
        onCancel,
    } = props

    const isEditMode = mode === 'edit'

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            ref={innerRef}
            noValidate={true}
            className={classnames(className, 'd-flex flex-column')}
            onSubmit={handleSubmit}
        >
            <FormInput
                required={true}
                autoFocus={true}
                title="Repository"
                description="This insight is limited to one repository. You can set up multiple language usage charts for analyzing other repositories."
                placeholder="Add or search for repository"
                valid={repository.meta.touched && repository.meta.validState === 'VALID'}
                error={repository.meta.touched && repository.meta.error}
                {...repository.input}
                className="mb-0"
            />

            <FormInput
                required={true}
                title="Title"
                description="Shown as the title for your insight."
                placeholder="Example: Migration to React function components"
                valid={title.meta.touched && title.meta.validState === 'VALID'}
                error={title.meta.touched && title.meta.error}
                {...title.input}
                className="mb-0 mt-4"
            />

            <FormInput
                required={true}
                min={1}
                max={100}
                type="number"
                title="Threshold of ‘Other’ category"
                description="Languages with usage lower than the threshold are grouped into an 'other' category."
                valid={threshold.meta.touched && threshold.meta.validState === 'VALID'}
                error={threshold.meta.touched && threshold.meta.error}
                {...threshold.input}
                className="mb-0 mt-4"
                inputClassName={styles.formThresholdInput}
                inputSymbol={<span className={styles.formThresholdInputSymbol}>%</span>}
            />

            <FormGroup
                name="visibility"
                title="Visibility"
                description="This insight will be visible only on your personal dashboard. It will not appear for other
                            users in your organization."
                className="mb-0 mt-4"
                contentClassName="d-flex flex-wrap mb-n2"
            >
                <FormRadioInput
                    name="visibility"
                    value="personal"
                    title="Personal"
                    description="only you"
                    checked={visibility.input.value === 'personal'}
                    className="mr-3"
                    onChange={visibility.input.onChange}
                />

                <FormRadioInput
                    name="visibility"
                    value="organization"
                    title="Organization"
                    description="all users in your organization"
                    checked={visibility.input.value === 'organization'}
                    onChange={visibility.input.onChange}
                    className="mr-3"
                />
            </FormGroup>

            <hr className={styles.formSeparator} />

            <div>
                {submitErrors?.[FORM_ERROR] && <ErrorAlert error={submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={submitting}
                    label={submitting ? 'Submitting' : isEditMode ? 'Edit insight' : 'Create code insight'}
                    type="submit"
                    disabled={submitting}
                    className="btn btn-primary mr-2"
                />

                <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>
                    Cancel
                </button>
            </div>
        </form>
    )
}
