import { FunctionComponent, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { isEqual } from 'lodash'
import PlusIcon from 'mdi-react/PlusIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Icon, Link, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../../components/LoaderButton'
import { useField } from '../../../../../../form/hooks/useField'
import { FORM_ERROR, FormChangeEvent, SubmissionResult, useForm } from '../../../../../../form/hooks/useForm'
import { DrillDownInput, LabelWithReset } from '../drill-down-input/DrillDownInput'
import { FilterCollapseSection } from '../filter-collapse-section/FilterCollapseSection'
import { DrillDownSearchContextFilter } from '../search-context/DrillDownSearchContextFilter'

import { getSerializedRepositoriesFilter, getSerializedSearchContextFilter, validRegexp } from './utils'
import { createSearchContextValidator, getFilterInputStatus } from './validators'

import styles from './DrillDownInsightFilters.module.scss'

enum FilterSection {
    SearchContext,
    RegularExpressions,
}

export enum FilterSectionVisualMode {
    CollapseSections,
    HorizontalSections,
}

export interface DrillDownFiltersFormValues {
    context: string
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

interface DrillDownInsightFilters {
    initialValues: DrillDownFiltersFormValues

    originalValues: DrillDownFiltersFormValues

    visualMode: FilterSectionVisualMode

    className?: string

    /** Fires whenever the user changes filter value in any form input. */
    onFiltersChange: (filters: FormChangeEvent<DrillDownFiltersFormValues>) => void

    /** Fires whenever the user clicks the save/update filter button. */
    onFilterSave: (filters: DrillDownFiltersFormValues) => SubmissionResult

    /** Fires whenever the user clicks the create insight button. */
    onCreateInsightRequest: () => void
}

export const DrillDownInsightFilters: FunctionComponent<DrillDownInsightFilters> = props => {
    const {
        initialValues,
        originalValues,
        className,
        visualMode,
        onFiltersChange,
        onFilterSave,
        onCreateInsightRequest,
    } = props

    const [activeSection, setActiveSection] = useState<FilterSection | null>(FilterSection.RegularExpressions)

    const { ref, formAPI, handleSubmit, values } = useForm<DrillDownFiltersFormValues>({
        initialValues,
        onChange: onFiltersChange,
        onSubmit: onFilterSave,
    })

    const client = useApolloClient()
    const contextValidator = useMemo(() => createSearchContextValidator(client), [client])

    const contexts = useField({
        name: 'context',
        formApi: formAPI,
        validators: { async: contextValidator },
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

    const currentRepositoriesFilters = { include: includeRegex.input.value, exclude: excludeRegex.input.value }
    const hasFiltersChanged = !isEqual(originalValues, values)
    const hasAppliedFilters = hasActiveFilters(originalValues)

    const handleCollapseState = (section: FilterSection, opened: boolean): void => {
        if (!opened) {
            setActiveSection(null)
        } else {
            setActiveSection(section)
        }
    }

    const handleClear = (): void => {
        contexts.input.onChange('')
        includeRegex.input.onChange('')
        excludeRegex.input.onChange('')
    }

    const isHorizontalMode = visualMode === FilterSectionVisualMode.HorizontalSections

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit} className={className}>
            <header className={styles.header}>
                <Typography.H4 className={styles.heading}>Filter repositories</Typography.H4>

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

            <hr className={styles.headerSeparator} />

            <div className={classNames(styles.panels, { [styles.panelsHorizontalMode]: isHorizontalMode })}>
                <FilterCollapseSection
                    open={isHorizontalMode || activeSection === FilterSection.SearchContext}
                    title="Search context"
                    preview={getSerializedSearchContextFilter(contexts.input.value)}
                    hasActiveFilter={hasActiveUnaryFilter(contexts.input.value)}
                    className={styles.panel}
                    withSeparators={!isHorizontalMode}
                    onOpenChange={opened => handleCollapseState(FilterSection.SearchContext, opened)}
                >
                    <small className={styles.sectionDescription}>
                        Choose{' '}
                        <Link
                            to="/help/code_search/how-to/search_contexts#beta-query-based-search-contexts"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            query-based search context (beta)
                        </Link>{' '}
                        to change the scope of this insight.
                    </small>

                    <DrillDownSearchContextFilter
                        spellCheck={false}
                        autoComplete="off"
                        autoFocus={true}
                        className={styles.input}
                        status={getFilterInputStatus(contexts)}
                        {...contexts.input}
                    />
                </FilterCollapseSection>

                <FilterCollapseSection
                    open={isHorizontalMode || activeSection === FilterSection.RegularExpressions}
                    title="Regular expression"
                    preview={getSerializedRepositoriesFilter(currentRepositoriesFilters)}
                    hasActiveFilter={
                        hasActiveUnaryFilter(includeRegex.input.value) || hasActiveUnaryFilter(excludeRegex.input.value)
                    }
                    className={styles.panel}
                    withSeparators={!isHorizontalMode}
                    onOpenChange={opened => handleCollapseState(FilterSection.RegularExpressions, opened)}
                >
                    <small className={styles.sectionDescription}>
                        Use regular expression to change the scope of this insight.
                    </small>

                    <fieldset className={styles.regExpFilters}>
                        <LabelWithReset
                            text="Include repositories"
                            disabled={!includeRegex.input.value}
                            onReset={() => includeRegex.input.onChange('')}
                        >
                            <DrillDownInput
                                autoFocus={true}
                                prefix="repo:"
                                placeholder="regexp-pattern"
                                spellCheck={false}
                                className={styles.input}
                                status={getFilterInputStatus(includeRegex)}
                                {...includeRegex.input}
                            />
                        </LabelWithReset>

                        <LabelWithReset
                            text="Exclude repositories"
                            disabled={!excludeRegex.input.value}
                            onReset={() => excludeRegex.input.onChange('')}
                        >
                            <DrillDownInput
                                prefix="-repo:"
                                placeholder="regexp-pattern"
                                spellCheck={false}
                                className={styles.input}
                                status={getFilterInputStatus(excludeRegex)}
                                {...excludeRegex.input}
                            />
                        </LabelWithReset>
                    </fieldset>
                </FilterCollapseSection>
            </div>

            {isHorizontalMode && <hr />}

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
                            Default filters
                        </Link>{' '}
                        applied
                    </small>
                )}

                <div className={styles.buttons}>
                    <LoaderButton
                        alwaysShowLabel={true}
                        loading={formAPI.submitting}
                        label={getSubmitButtonText({ submitting: formAPI.submitting, hasAppliedFilters })}
                        type="submit"
                        disabled={!formAPI.valid || formAPI.submitting || !hasFiltersChanged}
                        variant="secondary"
                        size="sm"
                        outline={true}
                    />

                    <Button
                        data-testid="save-as-new-view-button"
                        type="button"
                        variant="secondary"
                        size="sm"
                        disabled={!hasFiltersChanged || !formAPI.valid}
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

export function hasActiveFilters(filters: DrillDownFiltersFormValues): boolean {
    const { excludeRepoRegexp, includeRepoRegexp, context } = filters

    return [excludeRepoRegexp, includeRepoRegexp, context].some(hasActiveUnaryFilter)
}

const hasActiveUnaryFilter = (filter: string): boolean => filter.trim() !== ''

interface SubmitButtonTextProps {
    submitting: boolean
    hasAppliedFilters: boolean
}

function getSubmitButtonText(input: SubmitButtonTextProps): string {
    const { submitting, hasAppliedFilters } = input

    return submitting
        ? hasAppliedFilters
            ? 'Updating'
            : 'Saving'
        : hasAppliedFilters
        ? 'Update default filters'
        : 'Save default filters'
}
