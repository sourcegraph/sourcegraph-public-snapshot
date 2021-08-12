import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, PropsWithChildren, Ref } from 'react'

import { ErrorAlert } from '../../../../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../../../../components/LoaderButton'
import { TruncatedText } from '../../../../../../../../pages/dashboards/dashboard-page/components/dashboard-select/components/trancated-text/TrancatedText'
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useField } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, FormChangeEvent, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'
import { FlexTextArea } from '../../../../../../../form/repositories-field/components/flex-textarea/FlexTextArea'

import styles from './DrillDownFiltersForm.module.scss'
import { validRegexp } from './validators'

export interface DrillDownFiltersFormValues {
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

const INITIAL_FORM_VALUES: DrillDownFiltersFormValues = {
    includeRepoRegexp: '',
    excludeRepoRegexp: '',
}

interface DrillDownFiltersFormProps {
    className?: string
    initialFiltersValue?: DrillDownFiltersFormValues
    onFiltersChange: (filters: FormChangeEvent<DrillDownFiltersFormValues>) => void
    onFilterSave: (filters: DrillDownFiltersFormValues) => SubmissionResult
}

export const DrillDownFiltersForm: React.FunctionComponent<DrillDownFiltersFormProps> = props => {
    const { className, initialFiltersValue = INITIAL_FORM_VALUES, onFiltersChange, onFilterSave } = props

    const { ref, formAPI, handleSubmit } = useForm<DrillDownFiltersFormValues>({
        initialValues: initialFiltersValue,
        onChange: onFiltersChange,
        onSubmit: onFilterSave,
    })

    const includeRegex = useField({
        name: 'includeRepoRegexp',
        formApi: formAPI,
        validators: { sync: validRegexp },
    })

    const excludeRegex = useField({
        name: 'excludeRepoRegexp',
        formApi: formAPI,
        validators: { sync: validRegexp },
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} className={classnames(className, styles.drilldownFilters)} onSubmit={handleSubmit}>
            <header className="">
                <h4 className="mb-2">Filters by Repositories</h4>
                <p className="text-muted mb-2">
                    Default filters applied.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </header>

            <hr className="ml-n3 mr-n3" />

            <h4 className="mt-3 mb-3">Regular expression</h4>

            <FormInput
                as={DrillDownRegExpInput}
                autoFocus={true}
                prefix="repo:"
                title={
                    <LabelWithReset onReset={() => includeRegex.input.onChange('')}>
                        Include repositories
                    </LabelWithReset>
                }
                placeholder="^github\.com/sourcegraph/sourcegraph$"
                className="mb-4"
                valid={includeRegex.meta.dirty && includeRegex.meta.validState === 'VALID'}
                error={includeRegex.meta.dirty && includeRegex.meta.error}
                {...includeRegex.input}
            />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="-repo:"
                title={
                    <LabelWithReset onReset={() => excludeRegex.input.onChange('')}>
                        Exclude repositories
                    </LabelWithReset>
                }
                placeholder="^github\.com/sourcegraph/sourcegraph$"
                valid={excludeRegex.meta.dirty && excludeRegex.meta.validState === 'VALID'}
                error={excludeRegex.meta.dirty && excludeRegex.meta.error}
                className="mb-4"
                {...excludeRegex.input}
            />

            <hr className="ml-n3 mr-n3" />

            {formAPI.submitErrors?.[FORM_ERROR] && (
                <ErrorAlert className="mt-3 mb-3" error={formAPI.submitErrors[FORM_ERROR]} />
            )}

            <LoaderButton
                alwaysShowLabel={true}
                loading={formAPI.submitting}
                label={formAPI.submitting ? 'Updating' : 'Update default filters'}
                spinnerClassName="mr-2"
                type="submit"
                disabled={formAPI.submitting}
                className="d-flex btn btn-outline-secondary ml-auto mt-3 mb-1"
            />
        </form>
    )
}

interface LabelWithResetProps {
    onReset?: () => void
}

const LabelWithReset: React.FunctionComponent<PropsWithChildren<LabelWithResetProps>> = props => (
    <span className="d-flex align-items-center">
        <TruncatedText>{props.children}</TruncatedText>
        <button
            type="button"
            className="btn btn-link ml-auto pt-0 pb-0 pr-0 font-weight-normal"
            onClick={props.onReset}
        >
            Reset
        </button>
    </span>
)

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props

    return (
        <span className={classnames(styles.regexpField, 'w-100')}>
            <span className={styles.regexpFieldPrefix}>{prefix}</span>
            <FlexTextArea
                {...inputProps}
                className={classnames(inputProps.className, styles.regexpFieldInput)}
                ref={reference}
            />
        </span>
    )
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'
