import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import HelpCircleIcon from 'mdi-react/HelpCircleIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'
import React, { useCallback, useMemo, useRef, useState } from 'react'
import { Tooltip } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { FilterType, resolveFilter, validateFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { deriveInputClassName, useInputValidation } from '@sourcegraph/shared/src/util/useInputValidation'

import { SearchPatternType } from '../../../graphql-operations'

interface TriggerAreaProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: (complete: boolean) => void
    startExpanded: boolean
}

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'
const isLiteralOrRegexp = (value: string): boolean => value === 'literal' || value === 'regexp'

const ValidQueryChecklistItem: React.FunctionComponent<{ checked: boolean; hint?: string; className?: string }> = ({
    checked,
    children,
    hint,
    className,
}) => {
    const tooltipTarget = useRef<HTMLElement | null>(null)
    const [tooltipOpen, setTooltipOpen] = useState(false)
    const toggleTooltip = useCallback(() => setTooltipOpen(isOpen => !isOpen), [])
    const showTooltip = useCallback(() => setTooltipOpen(true), [])
    const hideTooltip = useCallback(() => setTooltipOpen(false), [])

    return (
        <label
            className={classNames('d-flex align-items-center mb-1 text-muted', className)}
            onMouseOver={showTooltip}
            onMouseOut={hideTooltip}
            onFocus={showTooltip}
            onBlur={hideTooltip}
        >
            <input className="sr-only" type="checkbox" disabled={true} checked={checked} />

            {checked ? (
                <CheckIcon className="trigger-area__checklist-checkbox icon-inline text-success" aria-hidden={true} />
            ) : (
                <RadioboxBlankIcon
                    className="trigger-area__checklist-checkbox trigger-area__checklist-checkbox--unchecked icon-inline"
                    aria-hidden={true}
                />
            )}

            <small className={checked ? 'trigger-area__checklist-children--faded' : ''}>{children}</small>

            {hint && (
                <>
                    <span className="sr-only"> {hint}</span>

                    <span ref={tooltipTarget} className="d-flex">
                        <HelpCircleIcon
                            className={classNames('trigger-area__checklist-hint', 'icon-inline', {
                                'trigger-area__checklist-hint--faded': checked,
                            })}
                            aria-hidden={true}
                        />
                    </span>

                    <Tooltip
                        target={tooltipTarget}
                        toggle={toggleTooltip}
                        isOpen={tooltipOpen}
                        placement="bottom"
                        innerClassName="trigger-area__checklist-tooltip"
                    >
                        {hint}
                    </Tooltip>
                </>
            )}
        </label>
    )
}

export const FormTriggerArea: React.FunctionComponent<TriggerAreaProps> = ({
    query,
    onQueryChange,
    triggerCompleted,
    setTriggerCompleted,
    startExpanded,
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
                <div className="code-monitor-form__card card p-3">
                    <div className="font-weight-bold">When there are new search results</div>
                    <span className="text-muted">
                        This trigger will fire when new search results are found for a given search query.
                    </span>
                    <span className="mt-4">Search query</span>
                    <div>
                        <div className="trigger-area__query-input mb-4">
                            <div className="d-flex flex-column flex-grow-1">
                                <input
                                    type="text"
                                    className={classNames(
                                        'trigger-area__query-input-field form-control mt-2 mb-3 test-trigger-input text-monospace',
                                        `test-${deriveInputClassName(queryState)}`
                                    )}
                                    onChange={nextQueryFieldChange}
                                    value={queryState.value}
                                    autoFocus={true}
                                    ref={queryInputReference}
                                    spellCheck={false}
                                    data-testid="trigger-query-edit"
                                />

                                <ul className="trigger-area__checklist">
                                    <li>
                                        <ValidQueryChecklistItem
                                            className="test-patterntype-checkbox"
                                            checked={hasValidPatternTypeFilter}
                                            hint="Code monitors support literal and regex search. Searches are literal by default."
                                        >
                                            Is <code>patternType:literal</code> or <code>patternType:regexp</code>
                                        </ValidQueryChecklistItem>
                                    </li>
                                    <li>
                                        <ValidQueryChecklistItem
                                            className="test-type-checkbox"
                                            checked={hasTypeDiffOrCommitFilter}
                                            hint="type:diff targets code present in new commits, while type:commit targets commit messages"
                                        >
                                            Contains a <code>type:diff</code> or <code>type:commit</code> filter
                                        </ValidQueryChecklistItem>
                                    </li>
                                    <li>
                                        <ValidQueryChecklistItem
                                            className="test-repo-checkbox"
                                            checked={hasRepoFilter}
                                            hint="Code monitors can watch a maximum of 50 repos at a time. Target your query with repo: filters to narrow down your search."
                                        >
                                            Contains a <code>repo:</code> filter
                                        </ValidQueryChecklistItem>
                                    </li>
                                    <li>
                                        <ValidQueryChecklistItem checked={isValidQuery} className="test-valid-checkbox">
                                            Is a valid search query
                                        </ValidQueryChecklistItem>
                                    </li>
                                </ul>
                            </div>
                            <div className="trigger-area__query-input-preview-link p-2 my-2">
                                <Link
                                    to={`/search?${buildSearchURLQuery(
                                        queryState.value,
                                        SearchPatternType.literal,
                                        false
                                    )}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="trigger-area__query-input-preview-link-text test-preview-link"
                                >
                                    Preview results{' '}
                                    <OpenInNewIcon className="trigger-area__query-input-preview-link-icon ml-1 icon-inline" />
                                </Link>
                            </div>
                        </div>
                    </div>
                    <div>
                        <button
                            className="btn btn-secondary mr-1 test-submit-trigger"
                            onClick={completeForm}
                            type="submit"
                            disabled={queryState.kind !== 'VALID'}
                        >
                            Continue
                        </button>
                        <button type="button" className="btn btn-outline-secondary" onClick={cancelForm}>
                            Cancel
                        </button>
                    </div>
                </div>
            )}
            {!showQueryForm && (
                <button
                    type="button"
                    className="btn code-monitor-form__card--button card test-trigger-button"
                    aria-label="Edit trigger: When there are new search results"
                    onClick={toggleQueryForm}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div
                                className={classNames(
                                    'font-weight-bold',
                                    !triggerCompleted && 'code-monitor-form__card-link btn-link'
                                )}
                            >
                                When there are new search results
                            </div>
                            {triggerCompleted ? (
                                <code
                                    className="trigger-area__query-label text-break text-muted test-existing-query"
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
                        {triggerCompleted && <div className="btn-link">Edit</div>}
                    </div>
                </button>
            )}
            <small className="text-muted">
                {' '}
                What other events would you like to monitor?{' '}
                <a href="mailto:feedback@sourcegraph.com" target="_blank" rel="noopener">
                    Share feedback.
                </a>
            </small>
        </>
    )
}
