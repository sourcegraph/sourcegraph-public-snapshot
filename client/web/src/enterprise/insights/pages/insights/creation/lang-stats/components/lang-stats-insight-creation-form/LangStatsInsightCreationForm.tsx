import React, { FormEventHandler, RefObject, useMemo } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Input, useObservable } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../components/LoaderButton'
import { CodeInsightDashboardsVisibility } from '../../../../../../components/creation-ui-kit'
import { getDefaultInputProps } from '../../../../../../components/form/getDefaultInputProps'
import { useFieldAPI } from '../../../../../../components/form/hooks/useField'
import { FORM_ERROR, SubmissionErrors } from '../../../../../../components/form/hooks/useForm'
import { RepositoryField } from '../../../../../../components/form/repositories-field/RepositoryField'
import { LimitedAccessLabel } from '../../../../../../components/limited-access-label/LimitedAccessLabel'
import { Insight } from '../../../../../../core/types'
import { useUiFeatures } from '../../../../../../hooks/use-ui-features'
import { LangStatsCreationFormFields } from '../../types'

import styles from './LangStatsInsightCreationForm.module.scss'

export interface LangStatsInsightCreationFormProps {
    mode?: 'creation' | 'edit'
    innerRef: RefObject<any>
    handleSubmit: FormEventHandler
    submitErrors: SubmissionErrors
    submitting: boolean
    className?: string
    isFormClearActive?: boolean
    dashboardReferenceCount?: number

    title: useFieldAPI<LangStatsCreationFormFields['title']>
    repository: useFieldAPI<LangStatsCreationFormFields['repository']>
    threshold: useFieldAPI<LangStatsCreationFormFields['threshold']>
    insight?: Insight

    onCancel: () => void
    onFormReset: () => void
}

export const LangStatsInsightCreationForm: React.FunctionComponent<
    React.PropsWithChildren<LangStatsInsightCreationFormProps>
> = props => {
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
        isFormClearActive,
        dashboardReferenceCount,
        onCancel,
        onFormReset,
        insight,
    } = props

    const isEditMode = mode === 'edit'
    const { licensed, insight: insightFeatures } = useUiFeatures()

    const creationPermission = useObservable(
        useMemo(
            () =>
                isEditMode && insight
                    ? insightFeatures.getEditPermissions(insight)
                    : insightFeatures.getCreationPermissions(),
            [insightFeatures, isEditMode, insight]
        )
    )

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            ref={innerRef}
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

            {!licensed && !isEditMode && (
                <LimitedAccessLabel
                    message="Unlock Code Insights to create unlimited insights"
                    className="my-3 mt-n2"
                />
            )}

            <div className="d-flex flex-wrap align-items-center">
                {submitErrors?.[FORM_ERROR] && <ErrorAlert className="w-100" error={submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    alwaysShowLabel={true}
                    data-testid="insight-save-button"
                    loading={submitting}
                    label={submitting ? 'Submitting' : isEditMode ? 'Save insight' : 'Create code insight'}
                    type="submit"
                    disabled={submitting || !creationPermission?.available}
                    className="mr-2 mb-2"
                    variant="primary"
                />

                <Button type="button" variant="secondary" outline={true} className="mb-2 mr-auto" onClick={onCancel}>
                    Cancel
                </Button>

                <Button
                    type="reset"
                    variant="secondary"
                    outline={true}
                    disabled={!isFormClearActive}
                    className="border-0"
                >
                    Clear all fields
                </Button>
            </div>
        </form>
    )
}
