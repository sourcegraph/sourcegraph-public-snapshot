import { FC, FormEventHandler, ReactNode } from 'react'

import classNames from 'classnames'

import { Input } from '@sourcegraph/wildcard'

import {
    CodeInsightDashboardsVisibility,
    getDefaultInputProps,
    useFieldAPI,
    SubmissionErrors,
    RepositoryField,
} from '../../../../../../components'
import { LangStatsCreationFormFields } from '../../types'

import styles from './LangStatsInsightCreationForm.module.scss'

export interface LangStatsInsightCreationFormProps {
    handleSubmit: FormEventHandler
    submitErrors: SubmissionErrors
    submitting: boolean
    className?: string
    isFormClearActive: boolean
    dashboardReferenceCount?: number

    title: useFieldAPI<LangStatsCreationFormFields['title']>
    repository: useFieldAPI<LangStatsCreationFormFields['repository']>
    threshold: useFieldAPI<LangStatsCreationFormFields['threshold']>

    onFormReset: () => void
    children: (input: RenderPropertyInputs) => ReactNode
}

export interface RenderPropertyInputs {
    submitting: boolean
    submitErrors: SubmissionErrors
    isFormClearActive: boolean
}

export const LangStatsInsightCreationForm: FC<LangStatsInsightCreationFormProps> = props => {
    const {
        handleSubmit,
        submitErrors,
        submitting,
        className,
        title,
        repository,
        threshold,
        isFormClearActive,
        dashboardReferenceCount,
        onFormReset,
        children,
    } = props

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            noValidate={true}
            className={classNames(className, 'd-flex flex-column')}
            onSubmit={handleSubmit}
            onReset={onFormReset}
        >
            {/*
                a11y-ignore
                Rule: aria-allowed-role ARIA - role should be appropriate for the element
                Error occurs as a result of using `role=combobox` on `textarea` element.
             */}
            <Input
                as={RepositoryField}
                required={true}
                autoFocus={true}
                label="Repository"
                message="This insight is limited to one repository. You can set up multiple language usage charts for analyzing other repositories."
                placeholder="Example: github.com/sourcegraph/sourcegraph"
                {...getDefaultInputProps(repository)}
                className="mb-0"
                inputClassName="a11y-ignore"
            />

            <Input
                required={true}
                label="Title"
                message="Shown as the title for your insight."
                placeholder="Example: Language Usage in RepositoryName"
                {...getDefaultInputProps(title)}
                className="mb-0 mt-4"
            />

            <Input
                required={true}
                min={1}
                max={100}
                type="number"
                label="Threshold of ‘Other’ category"
                message="Languages with usage lower than the threshold are grouped into an 'other' category."
                {...getDefaultInputProps(threshold)}
                className="mb-0 mt-4"
                inputClassName={styles.formThresholdInput}
                inputSymbol={<span className={styles.formThresholdInputSymbol}>%</span>}
            />

            {!!dashboardReferenceCount && dashboardReferenceCount > 1 && (
                <CodeInsightDashboardsVisibility className="mt-5 mb-n1" dashboardCount={dashboardReferenceCount} />
            )}

            <hr className={styles.formSeparator} />

            {children({ submitting, submitErrors, isFormClearActive })}
        </form>
    )
}
