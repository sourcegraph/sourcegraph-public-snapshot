import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import classnames from 'classnames'
import React, { useState, useCallback, useMemo } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { FilterType } from '../../../../../shared/src/search/interactive/util'
import { buildSearchURLQuery } from '../../../../../shared/src/util/url'
import { useInputValidation, deriveInputClassName } from '../../../../../shared/src/util/useInputValidation'
import { SearchPatternType } from '../../../graphql-operations'
import { scanSearchQuery } from '../../../../../shared/src/search/query/scanner'
import { resolveFilter, validateFilter } from '../../../../../shared/src/search/query/filters'

interface TriggerAreaProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: (complete: boolean) => void
}

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'

export const FormTriggerArea: React.FunctionComponent<TriggerAreaProps> = ({
    query,
    onQueryChange,
    triggerCompleted,
    setTriggerCompleted,
}) => {
    const [showQueryForm, setShowQueryForm] = useState(false)
    const toggleQueryForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowQueryForm(show => !show)
    }, [])

    const [queryState, nextQueryFieldChange, queryInputReference, overrideState] = useInputValidation(
        useMemo(
            () => ({
                initialValue: query,
                synchronousValidators: [
                    (value: string) => {
                        const tokens = scanSearchQuery(value)
                        if (tokens.type === 'success') {
                            const filters = tokens.term.filter(token => token.type === 'filter')
                            const hasTypeDiffOrCommitFilter = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.type &&
                                    ((filter.value?.type === 'literal' &&
                                        filter.value &&
                                        isDiffOrCommit(filter.value.value)) ||
                                        (filter.value?.type === 'quoted' &&
                                            filter.value &&
                                            isDiffOrCommit(filter.value.quotedValue)))
                            )
                            const hasPatternTypeFilter = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                                    filter.value &&
                                    validateFilter(filter.field.value, filter.value)
                            )
                            if (hasTypeDiffOrCommitFilter && hasPatternTypeFilter) {
                                return undefined
                            }
                            if (!hasTypeDiffOrCommitFilter) {
                                return 'Code monitors require queries to specify either `type:commit` or `type:diff`.'
                            }
                            if (!hasPatternTypeFilter) {
                                return 'Code monitors require queries to specify a `patternType:` of literal, regexp, or structural.'
                            }
                        }
                        return 'Failed to parse query'
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
            onQueryChange(queryState.value)
        },
        [setTriggerCompleted, setShowQueryForm, onQueryChange, queryState]
    )

    const editForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowQueryForm(true)
        },
        [setShowQueryForm]
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
            {!showQueryForm && !triggerCompleted && (
                <button
                    type="button"
                    onClick={toggleQueryForm}
                    className="code-monitor-form__card--button card p-3 my-3 w-100 test-trigger-button"
                >
                    <div
                        onClick={toggleQueryForm}
                        className="code-monitor-form__card-link btn btn-link font-weight-bold p-0 text-left"
                    >
                        When there are new search results
                    </div>
                    <span className="text-muted">
                        This trigger will fire when new search results are found for a given search query.
                    </span>
                </button>
            )}
            {showQueryForm && (
                <div className="code-monitor-form__card card p-3 my-3">
                    <div className="font-weight-bold">When there are new search results</div>
                    <span className="text-muted">
                        This trigger will fire when new search results are found for a given search query.
                    </span>
                    <div className="create-monitor-page__query-input">
                        <input
                            type="text"
                            className={classnames(
                                'create-monitor-page__query-input-field form-control my-2 test-trigger-input',
                                deriveInputClassName(queryState)
                            )}
                            onChange={nextQueryFieldChange}
                            value={queryState.value}
                            required={true}
                            autoFocus={true}
                            ref={queryInputReference}
                        />
                        {queryState.kind === 'VALID' && (
                            <Link
                                to={buildSearchURLQuery(query, SearchPatternType.literal, false)}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="create-monitor-page__query-input-preview-link test-preview-link"
                            >
                                Preview results <OpenInNewIcon className="icon-inline" />
                            </Link>
                        )}
                        {queryState.kind === 'INVALID' && (
                            <small className="invalid-feedback mb-4 test-trigger-error">{queryState.reason}</small>
                        )}
                        {(queryState.kind === 'NOT_VALIDATED' || queryState.kind === 'VALID') && (
                            <div className="d-flex mb-4 flex-column">
                                <small className="text-muted">
                                    Code monitors only support <code className="bg-code">type:diff</code> and{' '}
                                    <code className="bg-code">type:commit</code> search queries.
                                </small>
                            </div>
                        )}
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
            {!showQueryForm && triggerCompleted && (
                <div className="code-monitor-form__card card p-3 my-3">
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">When there are new search results</div>
                            <code className="text-muted test-existing-query">{query}</code>
                        </div>
                        <div>
                            <button
                                type="button"
                                onClick={editForm}
                                className="btn btn-link p-0 text-left test-edit-trigger"
                            >
                                Edit
                            </button>
                        </div>
                    </div>
                </div>
            )}
            <small className="text-muted">
                {' '}
                What other events would you like to monitor? {/* TODO: populate link */}
                <a href="" target="_blank" rel="noopener">
                    {/* TODO: populate link */}
                    Share feedback.
                </a>
            </small>
        </>
    )
}
