import classNames from 'classnames'
import { isEqual } from 'lodash'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Link } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../../../components/LoaderButton'
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useField } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, FormChangeEvent, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'

import { DrillDownRegExpInput, LabelWithReset } from './components/drill-down-reg-exp-input/DrillDownRegExpInput'
import styles from './DrillDownFiltersForm.module.scss'
import { validRegexp } from './validators'

export interface DrillDownFiltersFormValues {
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

export const hasActiveFilters = (filters?: DrillDownFiltersFormValues): boolean => {
    if (!filters) {
        return false
    }

    // We don't have the repo list mode support yet
    return filters.excludeRepoRegexp.trim() !== '' || filters.includeRepoRegexp.trim() !== ''
}

interface DrillDownFiltersFormProps {
    className?: string

    /**
     * Insight filters value that are stored in setting subject with
     * insight configuration object, change whenever the user click
     * save/update default filters.
     */
    originalFiltersValue: DrillDownFiltersFormValues

    /**
     * Live filters that lives only in runtime memory and can be different
     * from originalFiltersValue of insight until the user syncs them by
     * save/update default filters.
     */
    initialFiltersValue: DrillDownFiltersFormValues

    /**
     * Fires whenever the user changes filter value in any form input.
     */
    onFiltersChange: (filters: FormChangeEvent<DrillDownFiltersFormValues>) => void

    /**
     * Fires whenever the user clicks the save/update filter button.
     */
    onFilterSave: (filters: DrillDownFiltersFormValues) => SubmissionResult

    /**
     * Fires whenever the user clicks the create insight button.
     */
    onCreateInsightRequest: () => void
}

export const DrillDownFiltersForm: React.FunctionComponent<DrillDownFiltersFormProps> = props => {
    const {
        className,
        initialFiltersValue,
        originalFiltersValue,
        onFiltersChange,
        onFilterSave,
        onCreateInsightRequest,
    } = props

    const { ref, formAPI, handleSubmit, values } = useForm<DrillDownFiltersFormValues>({
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

    const hasFiltersChanged = !isEqual(originalFiltersValue, values)
    const hasAppliedFilters = hasActiveFilters(originalFiltersValue)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} className={classNames(className, 'd-flex flex-column px-3')} onSubmit={handleSubmit}>
            <header className={styles.header}>
                <h4 className="mb-0">Filter repositories</h4>

                {hasAppliedFilters && (
                    <small className="ml-auto">
                        <span className="text-muted">Default filters applied</span>{' '}
                        <Link
                            to="/help/code_insights/explanations/code_insights_filters"
                            target="_blank"
                            rel="noopener"
                            className="small"
                        >
                            Learn more.
                        </Link>
                    </small>
                )}
            </header>

            <hr className={styles.separator} />

            <small className={styles.description}>Use regular expression to change the scope of this insight.</small>

            <fieldset>
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
                    className="mb-3"
                    valid={includeRegex.meta.dirty && includeRegex.meta.validState === 'VALID'}
                    error={includeRegex.meta.dirty && includeRegex.meta.error}
                    spellCheck={false}
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
                    spellCheck={false}
                    className="mb-2"
                    {...excludeRegex.input}
                />
            </fieldset>

            <footer className={styles.footer}>
                {formAPI.submitErrors?.[FORM_ERROR] && (
                    <ErrorAlert className="w-100 mb-3" error={formAPI.submitErrors[FORM_ERROR]} />
                )}

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={
                        formAPI.submitting
                            ? hasAppliedFilters
                                ? 'Updating'
                                : 'Saving'
                            : hasAppliedFilters
                            ? 'Update default filters'
                            : 'Save default filters'
                    }
                    type="submit"
                    disabled={formAPI.submitting || !hasFiltersChanged}
                    className="ml-auto mr-2"
                    variant="secondary"
                    outline={true}
                />

                <Button
                    data-testid="save-as-new-view-button"
                    type="button"
                    variant="secondary"
                    onClick={onCreateInsightRequest}
                >
                    <PlusIcon className="icon-inline mr-1" />
                    Save as new view
                </Button>
            </footer>
        </form>
    )
}
