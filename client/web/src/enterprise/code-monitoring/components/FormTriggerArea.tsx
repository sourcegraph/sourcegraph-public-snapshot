import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import HelpCircleIcon from 'mdi-react/HelpCircleIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { FilterType, resolveFilter, validateFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { deriveInputClassName, useInputValidation } from '@sourcegraph/shared/src/util/useInputValidation'
import { Button, Link, Card } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'

import styles from './FormTriggerArea.module.scss'

interface TriggerAreaProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: (complete: boolean) => void
    startExpanded: boolean
    cardClassName?: string
    cardBtnClassName?: string
    cardLinkClassName?: string
}

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'
const isLiteralOrRegexp = (value: string): boolean => value === 'literal' || value === 'regexp'

const ValidQueryChecklistItem: React.FunctionComponent<{
    checked: boolean
    hint?: string
    className?: string
    dataTestid?: string
}> = ({ checked, children, hint, className, dataTestid }) => (
    <label className={classNames('d-flex align-items-center mb-1 text-muted', className)}>
        <input className="sr-only" type="checkbox" disabled={true} checked={checked} data-testid={dataTestid} />

        {checked ? (
            <CheckIcon
                className={classNames('icon-inline text-success', styles.checklistCheckbox)}
                aria-hidden={true}
            />
        ) : (
            <RadioboxBlankIcon
                className={classNames('icon-inline', styles.checklistCheckbox, styles.checklistCheckboxUnchecked)}
                aria-hidden={true}
            />
        )}

        <small className={checked ? styles.checklistChildrenFaded : ''}>{children}</small>

        {hint && (
            <>
                <span className="sr-only"> {hint}</span>

                <span data-tooltip={hint} data-placement="bottom" className="d-flex">
                    <HelpCircleIcon
                        className={classNames(
                            styles.checklistHint,
                            'icon-inline',
                            checked && styles.checklistHintFaded
                        )}
                        aria-hidden={true}
                    />
                </span>
            </>
        )}
    </label>
)

export const FormTriggerArea: React.FunctionComponent<TriggerAreaProps> = ({
    query,
    onQueryChange,
    triggerCompleted,
    setTriggerCompleted,
    startExpanded,
    cardClassName,
    cardBtnClassName,
    cardLinkClassName,
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

    const [queryState, nextQueryFieldChange, queryInputReference, overrideState] = useInputValidation(
        useMemo(
            () => ({
                initialValue: query,
                synchronousValidators: [
                    (value: string) => {
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

                        if (!isValidQuery) {
                            return 'Failed to parse query'
                        }

                        if (!hasTypeDiffOrCommitFilter) {
                            return 'Code monitors require queries to specify either `type:commit` or `type:diff`.'
                        }

                        if (!hasRepoFilter) {
                            return 'Code monitors require queries to specify a `repo:` filter.'
                        }

                        return undefined
                    },
                ],
            }),
            [query]
        )
    )

    const completeForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowQueryForm(false)
            setTriggerCompleted(true)
            onQueryChange(`${queryState.value}${hasPatternTypeFilter ? '' : ' patternType:literal'}`)
        },
        [setTriggerCompleted, setShowQueryForm, onQueryChange, queryState, hasPatternTypeFilter]
    )

    const cancelForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowQueryForm(false)
            overrideState({ value: query })
        },
        [setShowQueryForm, overrideState, query]
    )

    return (
        <>
            <h3>Trigger</h3>
            {showQueryForm && (
                <Card className={classNames(cardClassName, 'p-3')}>
                    <div className="font-weight-bold">When there are new search results</div>
                    <span className="text-muted">
                        This trigger will fire when new search results are found for a given search query.
                    </span>
                    <span className="mt-4">Search query</span>
                    <div>
                        <div className={classNames('mb-4', styles.queryInput)}>
                            <div className="d-flex flex-column flex-grow-1">
                                <input
                                    type="text"
                                    className={classNames(
                                        'form-control mt-2 mb-3 test-trigger-input text-monospace',
                                        styles.queryInputField,
                                        `test-${deriveInputClassName(queryState)}`
                                    )}
                                    onChange={nextQueryFieldChange}
                                    value={queryState.value}
                                    autoFocus={true}
                                    ref={queryInputReference}
                                    spellCheck={false}
                                    data-testid="trigger-query-edit"
                                />

                                <ul className={styles.checklist}>
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
                            <div className={classNames('p-2 my-2', styles.queryInputPreviewLink)}>
                                <Link
                                    to={`/search?${buildSearchURLQuery(
                                        queryState.value,
                                        SearchPatternType.literal,
                                        false
                                    )}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className={classNames('test-preview-link', styles.queryInputPreviewLinkText)}
                                >
                                    Preview results{' '}
                                    <OpenInNewIcon
                                        className={classNames('ml-1 icon-inline', styles.queryInputPreviewLinkIcon)}
                                    />
                                </Link>
                            </div>
                        </div>
                    </div>
                    <div>
                        <Button
                            data-testid="submit-trigger"
                            className="mr-1 test-submit-trigger"
                            onClick={completeForm}
                            type="submit"
                            disabled={queryState.kind !== 'VALID'}
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
                                    !triggerCompleted && classNames(cardLinkClassName, styles.triggerLabel)
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
