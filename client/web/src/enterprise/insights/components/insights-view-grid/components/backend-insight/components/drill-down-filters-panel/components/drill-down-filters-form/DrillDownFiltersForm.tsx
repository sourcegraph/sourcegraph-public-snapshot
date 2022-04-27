import React from 'react'

import classNames from 'classnames'
import { isEqual } from 'lodash'
import PlusIcon from 'mdi-react/PlusIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Link, Icon, Input } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../../../components/LoaderButton'
import { getDefaultInputProps } from '../../../../../../../form/getDefaultInputProps'
import { useField } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, FormChangeEvent, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'

import {
    DrillDownContextInput,
    DrillDownRegExpInput,
    LabelWithReset,
} from './components/drill-down-reg-exp-input/DrillDownRegExpInput'
import { validRegexp } from './validators'

import styles from './DrillDownFiltersForm.module.scss'

export interface DrillDownFiltersFormValues {
    contexts: string[]
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

export const hasActiveFilters = (filters: DrillDownFiltersFormValues): boolean => {
    const { excludeRepoRegexp, includeRepoRegexp, contexts } = filters

    // We don't have the repo list mode support yet
    return excludeRepoRegexp.trim() !== '' || includeRepoRegexp.trim() !== '' || contexts.length > 0
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
     * Live filters that live only in runtime memory and can be different
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

    const contexts = useField({
        name: 'contexts',
        formApi: formAPI,
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

    const handleClear = (): void => {
        contexts.input.onChange([])
        includeRegex.input.onChange('')
        excludeRegex.input.onChange('')
    }

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} className={classNames(className, styles.form)} onSubmit={handleSubmit}>
            <header className={styles.header}>
                <h4 className={styles.heading}>Filter repositories</h4>

                <Button
                    disabled={!hasActiveFilters(values)}
                    variant="link"
                    size="sm"
                    className={styles.clearFilters}
                    onClick={handleClear}
                >
                    Clear filters
                </Button>
            </header>

            <hr className={classNames(styles.separator, styles.separatorSmall)} />

            <small className={styles.description}>
                Use{' '}
                <Link to="/help/code_search/how-to/search_contexts#beta-query-based-search-contexts">
                    query-based search context (beta)
                </Link>{' '}
                or regular expression to change the scope of this insight.
            </small>

            <fieldset className={styles.fieldset}>
                <LabelWithReset
                    text="Search context"
                    disabled={!contexts.input.value.length}
                    onReset={() => contexts.input.onChange([])}
                >
                    <Input
                        as={DrillDownContextInput}
                        placeholder="global (default)"
                        spellCheck={false}
                        {...getDefaultInputProps(contexts)}
                    />
                </LabelWithReset>

                <LabelWithReset
                    text="Include repositories"
                    disabled={!includeRegex.input.value}
                    onReset={() => includeRegex.input.onChange('')}
                >
                    <Input
                        as={DrillDownRegExpInput}
                        autoFocus={true}
                        prefix="repo:"
                        placeholder="regexp-pattern"
                        spellCheck={false}
                        {...getDefaultInputProps(includeRegex)}
                    />
                </LabelWithReset>

                <LabelWithReset
                    text="Exclude repositories"
                    disabled={!excludeRegex.input.value}
                    onReset={() => excludeRegex.input.onChange('')}
                >
                    <Input
                        as={DrillDownRegExpInput}
                        prefix="-repo:"
                        placeholder="regexp-pattern"
                        spellCheck={false}
                        {...getDefaultInputProps(excludeRegex)}
                    />
                </LabelWithReset>
            </fieldset>

            <hr className={styles.separator} />

            <footer className={styles.footer}>
                {formAPI.submitErrors?.[FORM_ERROR] && (
                    <ErrorAlert className="w-100 mb-3" error={formAPI.submitErrors[FORM_ERROR]} />
                )}

                {hasAppliedFilters && (
                    <small className="text-muted">
                        <Link
                            to="/help/code_insights/explanations/code_insights_filters"
                            target="_blank"
                            rel="noopener"
                        >
                            Default filter
                        </Link>{' '}
                        applied
                    </small>
                )}

                <div className={styles.buttons}>
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
                        variant="secondary"
                        outline={true}
                    />

                    <Button
                        data-testid="save-as-new-view-button"
                        type="button"
                        variant="secondary"
                        disabled={!hasFiltersChanged}
                        onClick={onCreateInsightRequest}
                    >
                        <Icon className="mr-1" as={PlusIcon} />
                        Save as new view
                    </Button>
                </div>
            </footer>
        </form>
    )
}
