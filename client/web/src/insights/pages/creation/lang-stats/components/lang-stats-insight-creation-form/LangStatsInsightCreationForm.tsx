import classnames from 'classnames'
import React from 'react'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { ErrorAlert } from '../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../components/LoaderButton'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput'
import { useField } from '../../../../../components/form/hooks/useField'
import { FORM_ERROR, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator'
import { createRequiredValidator } from '../../../../../components/form/validators'
import { InsightTypeSuffix } from '../../../../../core/types'

import styles from './LangStatsInsightCreationForm.module.scss'

const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field for code insight.')
const thresholdFieldValidator = createRequiredValidator('Threshold is a required field for code insight.')

export interface LangStatsInsightCreationFormProps {
    settings: Settings | null
    className?: string
    onSubmit: (values: LangStatsCreationFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export interface LangStatsCreationFormFields {
    title: string
    repository: string
    threshold: number
    visibility: 'personal' | 'organization'
}

const INITIAL_VALUES: LangStatsCreationFormFields = {
    repository: '',
    title: '',
    threshold: 3,
    visibility: 'personal',
}

export const LangStatsInsightCreationForm: React.FunctionComponent<LangStatsInsightCreationFormProps> = props => {
    const { className, onSubmit, onCancel, settings } = props

    const { handleSubmit, formAPI, ref } = useForm<LangStatsCreationFormFields>({
        initialValues: INITIAL_VALUES,
        onSubmit,
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypeSuffix.langStats })

    const repository = useField('repository', formAPI, repositoriesFieldValidator)
    const title = useField('title', formAPI, titleValidator)
    const threshold = useField('threshold', formAPI, thresholdFieldValidator)
    const visibility = useField('visibility', formAPI)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            ref={ref}
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
                {formAPI.submitErrors?.[FORM_ERROR] && <ErrorAlert error={formAPI.submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Submitting' : 'Create code insight'}
                    type="submit"
                    disabled={formAPI.submitting}
                    className="btn btn-primary mr-2"
                />

                <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>
                    Cancel
                </button>
            </div>
        </form>
    )
}
