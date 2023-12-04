import { type FunctionComponent, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiArrowExpand, mdiArrowCollapse, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { isEqual, noop } from 'lodash'

import { SeriesSortDirection, SeriesSortMode } from '@sourcegraph/shared/src/graphql-operations'
import {
    Button,
    Icon,
    Link,
    H4,
    ErrorAlert,
    useField,
    type FormChangeEvent,
    type SubmissionResult,
    useForm,
    FORM_ERROR,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../../components/LoaderButton'
import { SortFilterSeriesPanel } from '../../sort-filter-series-panel/SortFilterSeriesPanel'
import { DrillDownInput, LabelWithReset } from '../drill-down-input/DrillDownInput'
import { FilterCollapseSection, FilterPreviewPill } from '../filter-collapse-section/FilterCollapseSection'
import { DrillDownSearchContextFilter } from '../search-context/DrillDownSearchContextFilter'

import {
    getSerializedRepositoriesFilter,
    getSerializedSearchContextFilter,
    getSerializedSortAndLimitFilter,
} from './utils'
import { createSearchContextValidator, getFilterInputStatus, REPO_FILTER_VALIDATORS } from './validators'

import styles from './DrillDownInsightFilters.module.scss'

enum FilterSection {
    SortFilter,
    SearchContext,
    RegularExpressions,
}

export enum FilterSectionVisualMode {
    CollapseSections,
    HorizontalSections,
    Preview,
}

export interface DrillDownFiltersFormValues {
    context: string
    includeRepoRegexp: string
    excludeRepoRegexp: string
    seriesDisplayOptions: {
        numSamples: number | null
        limit: number | null
        sortOptions: {
            mode: SeriesSortMode
            direction: SeriesSortDirection
        }
    }
}

interface DrillDownInsightFilters {
    initialValues: DrillDownFiltersFormValues

    originalValues: DrillDownFiltersFormValues

    visualMode: FilterSectionVisualMode

    isNumSamplesFilterAvailable: boolean

    className?: string

    /** Fires whenever the user changes filter value in any form input. */
    onFiltersChange: (filters: FormChangeEvent<DrillDownFiltersFormValues>) => void

    onFilterValuesChange?: (values: DrillDownFiltersFormValues) => void

    /** Fires whenever the user clicks the save/update filter button. */
    onFilterSave: (filters: DrillDownFiltersFormValues) => SubmissionResult

    /** Fires whenever the user clicks the create insight button. */
    onCreateInsightRequest: () => void

    onVisualModeChange?: (nextVisualMode: FilterSectionVisualMode) => void
}

export const DrillDownInsightFilters: FunctionComponent<DrillDownInsightFilters> = props => {
    const {
        initialValues,
        originalValues,
        className,
        visualMode,
        isNumSamplesFilterAvailable,
        onFiltersChange,
        onFilterSave,
        onCreateInsightRequest,
        onVisualModeChange = noop,
        onFilterValuesChange = noop,
    } = props

    const [activeSection, setActiveSection] = useState<FilterSection | null>(FilterSection.RegularExpressions)

    const { ref, formAPI, handleSubmit, values } = useForm<DrillDownFiltersFormValues>({
        initialValues,
        onChange: onFiltersChange,
        onPureValueChange: onFilterValuesChange,
        onSubmit: values => onFilterSave(values),
    })

    const client = useApolloClient()

    const contexts = useField({
        name: 'context',
        formApi: formAPI,
        validators: { async: useMemo(() => createSearchContextValidator(client), [client]) },
    })

    const includeRegex = useField({
        name: 'includeRepoRegexp',
        formApi: formAPI,
        validators: { sync: REPO_FILTER_VALIDATORS },
    })

    const excludeRegex = useField({
        name: 'excludeRepoRegexp',
        formApi: formAPI,
        validators: { sync: REPO_FILTER_VALIDATORS },
    })

    const seriesDisplayOptionsField = useField({
        name: 'seriesDisplayOptions',
        formApi: formAPI,
    })

    const currentRepositoriesFilters = { include: includeRegex.input.value, exclude: excludeRegex.input.value }
    const hasFiltersChanged = !isEqual(originalValues, values)
    const hasAppliedFilters = hasActiveFilters(originalValues) && !hasFiltersChanged

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
        seriesDisplayOptionsField.input.onChange({
            limit: null,
            numSamples: null,
            sortOptions: {
                mode: SeriesSortMode.RESULT_COUNT,
                direction: SeriesSortDirection.DESC,
            },
        })
    }

    const isHorizontalMode = visualMode === FilterSectionVisualMode.HorizontalSections
    const isPreviewMode = visualMode === FilterSectionVisualMode.Preview

    if (isPreviewMode) {
        return (
            <header className={classNames(className, styles.header)}>
                <H4 className={styles.heading}>Filters</H4>

                <FilterPreviewPill text={getSerializedSearchContextFilter(contexts.input.value, true)} />
                <FilterPreviewPill text={getSerializedRepositoriesFilter(currentRepositoriesFilters)} />
                <FilterPreviewPill text={getSerializedSortAndLimitFilter(seriesDisplayOptionsField.input.value)} />

                <Button
                    variant="link"
                    className={classNames(styles.actionButton, styles.actionButtonWithCollapsed)}
                    onClick={() => onVisualModeChange(FilterSectionVisualMode.HorizontalSections)}
                    aria-label="Open filters panel"
                >
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand} />
                </Button>
            </header>
        )
    }

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit} className={className}>
            <header className={styles.header}>
                <H4 className={classNames(styles.heading, styles.headingWithExpandedContent)}>Filters</H4>

                <Button
                    disabled={
                        !hasActiveFilters(values) &&
                        isEqual(originalValues.seriesDisplayOptions, values.seriesDisplayOptions)
                    }
                    variant="link"
                    size="sm"
                    className={styles.actionButton}
                    onClick={handleClear}
                >
                    Clear filters
                </Button>

                {isHorizontalMode && (
                    <Button
                        variant="link"
                        className={styles.actionButton}
                        onClick={() => onVisualModeChange(FilterSectionVisualMode.Preview)}
                        aria-label="Close filters panel"
                    >
                        <Icon aria-hidden={true} svgPath={mdiArrowCollapse} />
                    </Button>
                )}
            </header>
            <hr className={styles.headerSeparator} />

            <div className={classNames({ [styles.panelsHorizontalMode]: isHorizontalMode })}>
                <FilterCollapseSection
                    open={isHorizontalMode || activeSection === FilterSection.SortFilter}
                    title="Sort & Limit"
                    aria-label="sort and limit filter section"
                    preview={getSerializedSortAndLimitFilter(seriesDisplayOptionsField.input.value)}
                    hasActiveFilter={!isEqual(originalValues.seriesDisplayOptions, values.seriesDisplayOptions)}
                    withSeparators={!isHorizontalMode}
                    onOpenChange={opened => handleCollapseState(FilterSection.SortFilter, opened)}
                >
                    <SortFilterSeriesPanel
                        value={seriesDisplayOptionsField.input.value}
                        isNumSamplesFilterAvailable={isNumSamplesFilterAvailable}
                        onChange={seriesDisplayOptionsField.input.onChange}
                    />
                </FilterCollapseSection>

                <FilterCollapseSection
                    open={isHorizontalMode || activeSection === FilterSection.SearchContext}
                    title="Search context"
                    aria-label="search context filter section"
                    preview={getSerializedSearchContextFilter(contexts.input.value)}
                    hasActiveFilter={hasActiveUnaryFilter(contexts.input.value)}
                    withSeparators={!isHorizontalMode}
                    className={classNames(styles.panel, { [styles.panelHorizontalMode]: isHorizontalMode })}
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
                        autoFocus={!isHorizontalMode}
                        className={styles.input}
                        status={getFilterInputStatus(contexts)}
                        {...contexts.input}
                    />
                </FilterCollapseSection>

                <FilterCollapseSection
                    open={isHorizontalMode || activeSection === FilterSection.RegularExpressions}
                    title="Regular expression"
                    aria-label="regular expressions filter section"
                    preview={getSerializedRepositoriesFilter(currentRepositoriesFilters)}
                    hasActiveFilter={
                        hasActiveUnaryFilter(includeRegex.input.value) || hasActiveUnaryFilter(excludeRegex.input.value)
                    }
                    withSeparators={!isHorizontalMode}
                    className={classNames(styles.panel, { [styles.panelHorizontalMode]: isHorizontalMode })}
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
                        <Icon aria-hidden={true} className="mr-1" svgPath={mdiPlus} />
                        Save as new view
                    </Button>
                </div>
            </footer>
        </form>
    )
}

export function hasActiveFilters(filters: DrillDownFiltersFormValues): boolean {
    const { excludeRepoRegexp, includeRepoRegexp, context, seriesDisplayOptions } = filters

    const { numSamples, sortOptions, limit } = seriesDisplayOptions
    const hasDisplayOptionChanged =
        numSamples !== null ||
        limit !== null ||
        sortOptions.mode !== SeriesSortMode.RESULT_COUNT ||
        sortOptions.direction !== SeriesSortDirection.DESC

    return [excludeRepoRegexp, includeRepoRegexp, context].some(hasActiveUnaryFilter) || hasDisplayOptionChanged
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
