import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import HelpCircleIcon from 'mdi-react/HelpCircleIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'

import { QueryState } from '@sourcegraph/search'
import { LazyMonacoQueryInput } from '@sourcegraph/search-ui'
import { FilterType, resolveFilter, validateFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link, Card, Icon, Checkbox, Typography } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'
import { useExperimentalFeatures } from '../../../stores'

import styles from './FormTriggerArea.module.scss'

interface TriggerAreaProps extends ThemeProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: (complete: boolean) => void
    startExpanded: boolean
    cardClassName?: string
    cardBtnClassName?: string
    cardLinkClassName?: string
    isSourcegraphDotCom: boolean
}

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'
const isLiteralOrRegexp = (value: string): boolean => value === 'literal' || value === 'regexp'

const ValidQueryChecklistItem: React.FunctionComponent<
    React.PropsWithChildren<{
        checked: boolean
        hint?: string
        className?: string
        dataTestid?: string
    }>
> = ({ checked, children, hint, className, dataTestid }) => (
    <Checkbox
        wrapperClassName={classNames('d-flex align-items-center text-muted pl-0', className)}
        className="sr-only"
        disabled={true}
        checked={checked}
        data-testid={dataTestid}
        id={dataTestid || 'ValidQueryCheckListInput'}
        label={
            <div className="d-flex align-items-center mb-1">
                {checked ? (
                    <Icon
                        className={classNames('text-success', styles.checklistCheckbox)}
                        aria-hidden={true}
                        as={CheckIcon}
                    />
                ) : (
                    <Icon
                        className={classNames(styles.checklistCheckbox, styles.checklistCheckboxUnchecked)}
                        aria-hidden={true}
                        as={RadioboxBlankIcon}
                    />
                )}

                <small className={checked ? styles.checklistChildrenFaded : ''}>{children}</small>

                {hint && (
                    <>
                        <span className="sr-only"> {hint}</span>

                        <span data-tooltip={hint} data-placement="bottom" className="d-inline-flex">
                            <Icon
                                className={classNames(styles.checklistHint, checked && styles.checklistHintFaded)}
                                aria-hidden={true}
                                as={HelpCircleIcon}
                            />
                        </span>
                    </>
                )}
            </div>
        }
    />
)

export const FormTriggerArea: React.FunctionComponent<React.PropsWithChildren<TriggerAreaProps>> = ({
    query,
    onQueryChange,
    triggerCompleted,
    setTriggerCompleted,
    startExpanded,
    cardClassName,
    cardBtnClassName,
    cardLinkClassName,
    isLightTheme,
    isSourcegraphDotCom,
}) => {
    const [showQueryForm, setShowQueryForm] = useState(startExpanded)
    const toggleQueryForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowQueryForm(show => !show)
    }, [])

    const [isValidQuery, setIsValidQuery] = useState(false)
    const [hasTypeDiffOrCommitFilter, setHasTypeDiffOrCommitFilter] = useState(false)
    const [hasRepoFilter, setHasRepoFilter] = useState(false)
    const [hasPatternTypeFilter, setHasPatternTypeFilter] = useState(false)
    const [hasValidPatternTypeFilter, setHasValidPatternTypeFilter] = useState(true)
    const isTriggerQueryComplete = useMemo(
        () => isValidQuery && hasTypeDiffOrCommitFilter && hasRepoFilter && hasValidPatternTypeFilter,
        [hasRepoFilter, hasTypeDiffOrCommitFilter, hasValidPatternTypeFilter, isValidQuery]
    )

    const [queryState, setQueryState] = useState<QueryState>({ query: query || '' })

    const editorComponent = useExperimentalFeatures(features => features.editor ?? 'monaco')

    useEffect(() => {
        const value = queryState.query
        const tokens = scanSearchQuery(value)

        const isValidQuery = !!value && tokens.type === 'success'
        setIsValidQuery(isValidQuery)

        let hasTypeDiffOrCommitFilter = false
        let hasRepoFilter = false
        let hasPatternTypeFilter = false
        let hasValidPatternTypeFilter = true

        if (tokens.type === 'success') {
            const filters = tokens.term.filter(token => token.type === 'filter')
            hasTypeDiffOrCommitFilter = filters.some(
                filter =>
                    filter.type === 'filter' &&
                    resolveFilter(filter.field.value)?.type === FilterType.type &&
                    filter.value &&
                    isDiffOrCommit(filter.value.value)
            )

            hasRepoFilter = filters.some(
                filter =>
                    filter.type === 'filter' &&
                    resolveFilter(filter.field.value)?.type === FilterType.repo &&
                    filter.value
            )

            hasPatternTypeFilter = filters.some(
                filter =>
                    filter.type === 'filter' &&
                    resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                    filter.value &&
                    validateFilter(filter.field.value, filter.value)
            )

            // No explicit patternType filter means we default
            // to patternType:literal
            hasValidPatternTypeFilter =
                !hasPatternTypeFilter ||
                filters.some(
                    filter =>
                        filter.type === 'filter' &&
                        resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                        filter.value &&
                        isLiteralOrRegexp(filter.value.value)
                )
        }

        setHasTypeDiffOrCommitFilter(hasTypeDiffOrCommitFilter)
        setHasRepoFilter(hasRepoFilter)
        setHasPatternTypeFilter(hasPatternTypeFilter)
        setHasValidPatternTypeFilter(hasValidPatternTypeFilter)
    }, [queryState.query])

    const completeForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowQueryForm(false)
            setTriggerCompleted(true)
            onQueryChange(`${queryState.query}${hasPatternTypeFilter ? '' : ' patternType:literal'}`)
        },
        [setTriggerCompleted, setShowQueryForm, onQueryChange, queryState, hasPatternTypeFilter]
    )

    const cancelForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowQueryForm(false)
            setQueryState({ query })
        },
        [setShowQueryForm, query]
    )

    const derivedInputClassName = useMemo(() => {
        if (!queryState.query) {
            return ''
        }
        if (isTriggerQueryComplete) {
            return 'is-valid'
        }
        return 'is-invalid'
    }, [isTriggerQueryComplete, queryState.query])

    return (
        <>
            <Typography.H3 as={Typography.H2}>Trigger</Typography.H3>
            {showQueryForm && (
                <Card className={classNames(cardClassName, 'p-3')}>
                    <div className="font-weight-bold">When there are new search results</div>
                    <span className="text-muted">
                        This trigger will fire when new search results are found for a given search query.
                    </span>
                    <span className="mt-4">Search query</span>
                    <div>
                        <div className={classNames(styles.queryInput, 'my-2')}>
                            <div
                                className={classNames(
                                    'form-control',
                                    styles.queryInputField,
                                    'test-trigger-input',
                                    `test-${derivedInputClassName}`
                                )}
                                data-testid="trigger-query-edit"
                            >
                                <LazyMonacoQueryInput
                                    editorComponent={editorComponent}
                                    isLightTheme={isLightTheme}
                                    patternType={SearchPatternType.literal}
                                    isSourcegraphDotCom={isSourcegraphDotCom}
                                    caseSensitive={false}
                                    queryState={queryState}
                                    onChange={setQueryState}
                                    onSubmit={() => {}}
                                    globbing={false}
                                    preventNewLine={false}
                                    autoFocus={true}
                                />
                            </div>
                            <div className={styles.queryInputPreviewLink}>
                                <Link
                                    to={`/search?${buildSearchURLQuery(
                                        queryState.query,
                                        SearchPatternType.literal,
                                        false
                                    )}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="test-preview-link"
                                >
                                    Preview results{' '}
                                    <Icon
                                        className={classNames('ml-1', styles.queryInputPreviewLinkIcon)}
                                        as={OpenInNewIcon}
                                    />
                                </Link>
                            </div>
                        </div>

                        <ul className={classNames(styles.checklist, 'mb-4')}>
                            <li>
                                <ValidQueryChecklistItem
                                    checked={hasValidPatternTypeFilter}
                                    hint="Code monitors support literal and regex search. Searches are literal by default."
                                    dataTestid="patterntype-checkbox"
                                >
                                    Is <code>patternType:literal</code> or <code>patternType:regexp</code>
                                </ValidQueryChecklistItem>
                            </li>
                            <li>
                                <ValidQueryChecklistItem
                                    checked={hasTypeDiffOrCommitFilter}
                                    hint="type:diff targets code present in new commits, while type:commit targets commit messages"
                                    dataTestid="type-checkbox"
                                >
                                    Contains a <code>type:diff</code> or <code>type:commit</code> filter
                                </ValidQueryChecklistItem>
                            </li>
                            <li>
                                <ValidQueryChecklistItem
                                    checked={hasRepoFilter}
                                    hint="Code monitors can watch a maximum of 50 repos at a time. Target your query with repo: filters to narrow down your search."
                                    dataTestid="repo-checkbox"
                                >
                                    Contains a <code>repo:</code> filter
                                </ValidQueryChecklistItem>
                            </li>
                            <li>
                                <ValidQueryChecklistItem checked={isValidQuery} dataTestid="valid-checkbox">
                                    Is a valid search query
                                </ValidQueryChecklistItem>
                            </li>
                        </ul>
                    </div>
                    <div>
                        <Button
                            data-testid="submit-trigger"
                            className="mr-1 test-submit-trigger"
                            onClick={completeForm}
                            type="submit"
                            disabled={!isTriggerQueryComplete}
                            variant="secondary"
                        >
                            Continue
                        </Button>
                        <Button onClick={cancelForm} outline={true} variant="secondary">
                            Cancel
                        </Button>
                    </div>
                </Card>
            )}
            {!showQueryForm && (
                <Card
                    data-testid="trigger-button"
                    as={Button}
                    className={classNames('test-trigger-button', cardBtnClassName)}
                    aria-label="Edit trigger: When there are new search results"
                    onClick={toggleQueryForm}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div
                                className={classNames(
                                    'font-weight-bold',
                                    triggerCompleted
                                        ? styles.triggerBtnText
                                        : classNames(cardLinkClassName, styles.triggerLabel)
                                )}
                            >
                                When there are new search results
                            </div>
                            {triggerCompleted ? (
                                <code
                                    className={classNames('text-break text-muted', styles.queryLabel)}
                                    data-testid="trigger-query-existing"
                                >
                                    {query}
                                </code>
                            ) : (
                                <span className="text-muted">
                                    This trigger will fire when new search results are found for a given search query.
                                </span>
                            )}
                        </div>
                        {triggerCompleted && (
                            <Button variant="link" as="div">
                                Edit
                            </Button>
                        )}
                    </div>
                </Card>
            )}
            <small className="text-muted">
                {' '}
                What other events would you like to monitor?{' '}
                <Link to="mailto:feedback@sourcegraph.com" target="_blank" rel="noopener">
                    Share feedback.
                </Link>
            </small>
        </>
    )
}
