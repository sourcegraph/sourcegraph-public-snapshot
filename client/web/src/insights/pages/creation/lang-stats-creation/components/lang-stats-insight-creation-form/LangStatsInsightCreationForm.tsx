import classnames from 'classnames';
import { FORM_ERROR, FormApi, SubmissionErrors } from 'final-form';
import createFocusDecorator from 'final-form-focus';
import React, { useEffect, useMemo, useRef } from 'react';
import { useField, useForm } from 'react-final-form-hooks';
import { noop } from 'rxjs';

import { InputField } from '../../../../../components/form/form-field/FormField';
import { FormGroup } from '../../../../../components/form/form-group/FormGroup';
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput';
import { composeValidators, createRequiredValidator } from '../../../../../components/form/validators';

import styles from './LangStatsInsightCreationForm.module.scss';

const requiredTitleField = createRequiredValidator('Title is required field for code insight.')
const repositoriesFieldValidator = composeValidators(
    createRequiredValidator('Repositories is required field for code insight.')
)

export interface LangStatsInsightCreationFormProps {
    className?: string;
    onSubmit: (
        values: LangStatsCreationFormFields,
        form: FormApi<LangStatsCreationFormFields, Partial<LangStatsCreationFormFields>>
    ) => SubmissionErrors | Promise<SubmissionErrors> | void
}

export interface LangStatsCreationFormFields {
    title: string;
    repository: string;
    threshold: number;
    visibility: 'personal' | 'organisation'
}

const INITIAL_VALUES: Partial<LangStatsCreationFormFields> = {
    threshold: 2,
    visibility: 'personal'
}

export const LangStatsInsightCreationForm: React.FunctionComponent<LangStatsInsightCreationFormProps> = props => {
    const { className, onSubmit } = props;

    const titleReference = useRef<HTMLInputElement>(null);
    const repositoryReference = useRef<HTMLInputElement>(null);
    const thresholdReference = useRef<HTMLInputElement>(null);

    const focusOnErrorsDecorator = useMemo(() => {
        const noopFocus = { focus: noop, name: '' }

        return createFocusDecorator<LangStatsCreationFormFields>(() => [
            repositoryReference.current ?? noopFocus,
            titleReference.current ?? noopFocus,
            thresholdReference.current ?? noopFocus,
        ])
    }, [])

    const { handleSubmit, form, submitErrors } = useForm<LangStatsCreationFormFields>({
        initialValues: INITIAL_VALUES,
        onSubmit
    });

    const repository = useField('repository', form, repositoriesFieldValidator);
    const title = useField('title', form, requiredTitleField);
    const threshold = useField('threshold', form)
    const visibility = useField('visibility', form);

    useEffect(() => focusOnErrorsDecorator(form), [form, focusOnErrorsDecorator])

    return (
        // eslint-disable-next-line react/forbid-elements
        <form className={classnames(className, styles.form)} onSubmit={handleSubmit}>

            <InputField
                autofocus={true}
                title="Repository"
                description="This insight is limited to one repository. You can set up muliple language usage charts for many repositories."
                placeholder="Add or search for repository"
                valid={repository.meta.touched && repository.meta.valid}
                error={repository.meta.touched && repository.meta.error}
                {...repository.input}
                ref={repositoryReference}
                className={styles.formField}
            />

            <InputField
                title="Title"
                description="Shown as title for your insight"
                placeholder="ex. Migration to React function components"
                valid={title.meta.touched && title.meta.valid}
                error={title.meta.touched && title.meta.error}
                {...title.input}
                ref={titleReference}
                className={styles.formField}
            />

            <InputField
                title="Threshold of ‘Other’ category"
                description="the threshold for grouping all other languages into an 'other' category"
                valid={threshold.meta.touched && threshold.meta.valid}
                error={threshold.meta.touched && threshold.meta.error}
                {...threshold.input}
                ref={thresholdReference}
                className={classnames(styles.formField)}
                inputClassName={styles.formThresholdInput}
                inputSymbol={<span className={styles.formThresholdInputSymbol}>%</span>}/>

            <FormGroup
                name="visibility"
                title="Visibility"
                description="This insight will be visible only on your personal dashboard. It will not be show to other
                            users in your organisation."
                className={styles.formField}
            >
                <div className={styles.formRadioGroupContent}>
                    <FormRadioInput
                        name="visibility"
                        value="personal"
                        title="Personal"
                        description="only for you"
                        checked={visibility.input.value === 'personal'}
                        className={styles.formRadio}
                        onChange={visibility.input.onChange}
                    />

                    <FormRadioInput
                        name="visibility"
                        value="organization"
                        title="Organization"
                        description="to all users in your organization"
                        checked={visibility.input.value === 'organization'}
                        onChange={visibility.input.onChange}
                        className={styles.formRadio}
                    />
                </div>
            </FormGroup>

            <div className={styles.formButtons}>
                {submitErrors?.[FORM_ERROR] && (
                    <div className="alert alert-danger">{submitErrors[FORM_ERROR].toString()}</div>
                )}

                <button
                    type="submit"
                    className={classnames(styles.formButton, 'btn btn-primary')}
                >
                    Create code insight
                </button>
                <button type="button" className={classnames(styles.formButton, 'btn btn-outline-secondary')}>
                    Cancel
                </button>
            </div>
        </form>
    );
}
