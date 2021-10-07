import classnames from 'classnames'
import { isEqual } from 'lodash'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../../../../../components/LoaderButton'
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useField } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, FormChangeEvent, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'

import { DrillDownRegExpInput, LabelWithReset } from './components/drill-down-reg-exp-input/DrillDownRegExpInput'
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
        <form ref={ref} className={classnames(className, 'd-flex flex-column')} onSubmit={handleSubmit}>
            <header className="d-flex align-items-baseline px-3 py-2 mt-1">
                <h4 className="mb-0">Filter repositories</h4>

                {hasAppliedFilters && (
                    <small className="text-muted ml-auto mb-0">
                        Default filters applied.{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_insights/explanations/code_insights_filters"
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more.
                        </a>
                    </small>
                )}
            </header>

            <hr className="w-100 m-0 mt-1" />

            <fieldset className="px-3 mt-3">
                <h4 className="mb-3">Regular expression</h4>

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
            </fieldset>

            <hr className="w-100 m-0" />

            <footer className="px-3 d-flex flex-wrap py-3">
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
                    className="btn btn-outline-secondary ml-auto mr-2"
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
