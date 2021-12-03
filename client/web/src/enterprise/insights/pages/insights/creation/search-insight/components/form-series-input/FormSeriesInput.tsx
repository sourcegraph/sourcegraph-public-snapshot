import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'
import Info from 'mdi-react/InfoCircleOutlineIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'
import React, { useMemo, useState } from 'react'
import { noop } from 'rxjs'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { useInputValidation } from '@sourcegraph/shared/src/util/useInputValidation'
import { Button } from '@sourcegraph/wildcard/src'

import { FormInput } from '../../../../../../components/form/form-input/FormInput'
import { useField } from '../../../../../../components/form/hooks/useField'
import { useForm } from '../../../../../../components/form/hooks/useForm'
import { MonacoField } from '../../../../../../components/form/monaco-field/MonacoField'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { SearchBasedInsightSeries } from '../../../../../../core/types/insight/search-insight'
import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput'

import styles from './FormSeriesInput.module.scss'

const requiredNameField = createRequiredValidator('Name is a required field for data series.')
const validQuery = createRequiredValidator('Query is a required field for data series.')

interface FormSeriesInputProps {
    id: string | null

    /** Series index. */
    index: number

    /**
     * This prop represents the case whenever the edit insight UI page
     * deals with backend insight. We need to disable our search insight
     * query field since our backend insight can't update BE data according
     * to the latest insight configuration.
     */
    isSearchQueryDisabled: boolean

    /**
     * Show all validation error of all fields within the form.
     */
    showValidationErrorsOnMount?: boolean
    /** Name of series. */
    name?: string
    /** Query value of series. */
    query?: string
    /** Color value for line chart. (series) */
    stroke?: string
    /** Enable autofocus behavior of first input of form. */
    autofocus?: boolean
    /** Enable cancel button. */
    cancel?: boolean
    /** Custom class name for root element of form series. */
    className?: string
    /** On submit handler of series form. */
    onSubmit?: (series: SearchBasedInsightSeries) => void
    /** On cancel handler. */
    onCancel?: () => void
    /** Change handler in order to listen last values of series form. */
    onChange?: (formValues: SearchBasedInsightSeries, valid: boolean) => void
}

const CheckListItem: React.FunctionComponent<{ valid?: boolean }> = ({ children, valid }) => {
    const StatusIcon: React.FunctionComponent = () =>
        valid ? (
            <Check size={16} className="text-success icon-inline" style={{ top: '3px' }} />
        ) : (
            <RadioboxBlankIcon size={16} className="icon-inline" style={{ top: '3px' }} />
        )
    return (
        <>
            <StatusIcon /> {children}
        </>
    )
}

interface SearchQueryChecksProps {
    checks: {
        isValidRegex: boolean
        isValidOperator: boolean
        isValidPatternType: boolean
        isNotRepoOrFile: boolean
        isNotCommitOrDiff: boolean
        isNoRepoFilter: boolean
    }
}
const SearchQueryChecks: React.FunctionComponent<SearchQueryChecksProps> = ({ checks }) => (
    <div className={classNames(styles.formSeriesInput)}>
        <ul className={classNames(['mt-4 text-muted', styles.formSeriesInputSeriesCheck])}>
            <li>
                <CheckListItem valid={checks.isValidRegex}>
                    Contains a properly formatted regular expression with at least one capture group
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidOperator}>
                    Does not contain boolean operator <code>AND</code> and <code>OR</code> (regular expression boolean
                    operators can still be used)
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidPatternType}>
                    Does not contain <code>patternType:literal</code> and <code>patternType:structural</code>
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotRepoOrFile}>
                    The capture group matches file contents (not <code>repo</code> or <code>file</code>)
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotCommitOrDiff}>
                    Does not contain <code>commit</code> or <code>diff</code> search
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNoRepoFilter}>
                    Does not contain the <code>repo:</code> filter as it will be added automatically if needed
                </CheckListItem>
            </li>
        </ul>
        <p className="mt-4 text-muted">
            Tip: use <code>archived:no</code> or <code>fork:no</code> to exclude results from archived or forked
            repositories. Explore{' '}
            <a href="https://docs.sourcegraph.com/code_insights/references/common_use_cases">example queries</a> and
            learn more about{' '}
            <a href="https://docs.sourcegraph.com/code_insights/references/common_reasons_code_insights_may_not_match_search_results">
                automatically generated data series
            </a>
            .
        </p>
        <p className="mt-4 text-muted">
            <Info size={16} /> <b>Name</b> and <b>color</b> of each data seris will be generated automatically. Chart
            will display <b>up to 20</b> data series.
        </p>
    </div>
)

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'

/** Displays form series input (three field - name field, query field and color picker). */
export const FormSeriesInput: React.FunctionComponent<FormSeriesInputProps> = props => {
    const {
        id,
        index,
        isSearchQueryDisabled,
        showValidationErrorsOnMount = false,
        name,
        query,
        stroke: color,
        className,
        cancel = false,
        autofocus = true,
        onCancel = noop,
        onSubmit = noop,
        onChange = noop,
    } = props

    const hasNameControlledValue = !!name
    const hasQueryControlledValue = !!query

    const { formAPI, handleSubmit, ref } = useForm({
        touched: showValidationErrorsOnMount,
        initialValues: {
            seriesName: name ?? '',
            seriesQuery: query ?? '',
            seriesColor: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit: values =>
            onSubmit({
                id,
                name: values.seriesName,
                query: values.seriesQuery,
                stroke: values.seriesColor,
            }),
        onChange: event => {
            const { values } = event

            onChange(
                {
                    id,
                    name: values.seriesName,
                    query: values.seriesQuery,
                    stroke: values.seriesColor,
                },
                event.valid
            )
        },
    })

    const nameField = useField({
        name: 'seriesName',
        formApi: formAPI,
        validators: { sync: requiredNameField },
    })

    const queryField = useField({
        name: 'seriesQuery',
        formApi: formAPI,
        validators: { sync: validQuery },
        disabled: isSearchQueryDisabled,
    })

    const colorField = useField({
        name: 'seriesColor',
        formApi: formAPI,
    })

    // Search query validators
    const [isValidRegex, setIsValidRegex] = useState(false)
    const [isValidOperator, setIsValidOperator] = useState(false)
    const [isValidPatternType, setIsValidPatternType] = useState(false)
    const [isNotRepoOrFile, setIsNotRepoOrFile] = useState(false)
    const [isNotCommitOrDiff, setIsNotCommitOrDiff] = useState(false)
    const [isNoRepoFilter, setIsNoRepoFilter] = useState(false)

    const [queryState, nextQueryFieldChange] = useInputValidation(
        useMemo(
            () => ({
                initialValue: queryField.input.value,
                synchronousValidators: [
                    (value: string) => {
                        const tokens = scanSearchQuery(value)

                        const validRegex = false
                        let validOperator = false
                        let validPatternType = false
                        const notRepoOrFile = false
                        let notCommitOrDiff = false
                        let noRepoFilter = false

                        if (tokens.type === 'success') {
                            const filters = tokens.term.filter(token => token.type === 'filter')

                            notCommitOrDiff = !filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.type &&
                                    filter.value &&
                                    isDiffOrCommit(filter.value.value)
                            )

                            noRepoFilter = !filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.repo &&
                                    filter.value
                            )

                            const hasLiteral = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                                    filter.value?.value === 'literal'
                            )

                            const hasStructural = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                                    filter.value?.value === 'structural'
                            )

                            validPatternType = !(hasLiteral && hasStructural)

                            const hasAnd = filters.some(filter => filter.type === 'keyword' && filter.value === 'AND')
                            const hasOr = filters.some(filter => filter.type === 'keyword' && filter.value === 'OR')

                            validOperator = !(hasAnd || hasOr)
                        }

                        setIsValidRegex(validRegex)
                        setIsValidOperator(validOperator)
                        setIsValidPatternType(validPatternType)
                        setIsNotRepoOrFile(notRepoOrFile)
                        setIsNotCommitOrDiff(notCommitOrDiff)
                        setIsNoRepoFilter(noRepoFilter)

                        return undefined
                    },
                ],
            }),
            [queryField.input.value]
        )
    )

    return (
        <div data-testid="series-form" ref={ref} className={classNames('d-flex flex-column', className)}>
            <FormInput
                title="Name"
                required={true}
                autoFocus={autofocus}
                placeholder="Example: Function component"
                description="Name shown in the legend and tooltip"
                valid={(hasNameControlledValue || nameField.meta.touched) && nameField.meta.validState === 'VALID'}
                error={nameField.meta.touched && nameField.meta.error}
                {...nameField.input}
            />

            <FormInput
                title="Search query"
                required={true}
                as={MonacoField}
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                description={
                    <span>
                        {!isSearchQueryDisabled ? (
                            <SearchQueryChecks
                                checks={{
                                    isValidRegex,
                                    isValidOperator,
                                    isValidPatternType,
                                    isNotRepoOrFile,
                                    isNotCommitOrDiff,
                                    isNoRepoFilter,
                                }}
                            />
                        ) : (
                            <>
                                We don't yet allow editing queries for insights over all repos. To change the query,
                                make a new insight. This is a known{' '}
                                <a
                                    href="https://docs.sourcegraph.com/code_insights/explanations/current_limitations_of_code_insights"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    beta limitation
                                </a>
                            </>
                        )}
                    </span>
                }
                valid={(hasQueryControlledValue || queryField.meta.touched) && queryField.meta.validState === 'VALID'}
                error={queryField.meta.touched && queryField.meta.error}
                className="mt-4"
                {...queryField.input}
            />

            <FormColorInput
                name={`color group of ${index} series`}
                title="Color"
                className="mt-4"
                value={colorField.input.value}
                onChange={colorField.input.onChange}
            />

            <div className="mt-4">
                <Button
                    aria-label="Submit button for data series"
                    type="button"
                    variant="secondary"
                    onClick={handleSubmit}
                >
                    Done
                </Button>

                {cancel && (
                    <Button type="button" onClick={onCancel} variant="secondary" outline={true} className="ml-2">
                        Cancel
                    </Button>
                )}
            </div>
        </div>
    )
}
