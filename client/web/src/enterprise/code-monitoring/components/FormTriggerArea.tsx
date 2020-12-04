import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import classnames from 'classnames'
import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { FilterType } from '../../../../../shared/src/search/interactive/util'
import { resolveFilter, validateFilter } from '../../../../../shared/src/search/parser/filters'
import { scanSearchQuery } from '../../../../../shared/src/search/parser/scanner'
import { buildSearchURLQuery } from '../../../../../shared/src/util/url'
import { useInputValidation, deriveInputClassName } from '../../../../../shared/src/util/useInputValidation'
import { SearchPatternType } from '../../../graphql-operations'

interface TriggerAreaProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: () => void
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

    const editOrCompleteForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            toggleQueryForm(event)
            setTriggerCompleted()
        },
        [setTriggerCompleted, toggleQueryForm]
    )

    const [queryState, nextQueryFieldChange, queryInputReference] = useInputValidation(
        useMemo(
            () => ({
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
            []
        )
    )

    useEffect(() => {
        if (queryState.kind === 'VALID') {
            onQueryChange(queryState.value)
        }
    }, [onQueryChange, queryState])

    return (
        <>
            <h3>Trigger</h3>
            <div className="card p-3 my-3">
                {!showQueryForm && !triggerCompleted && (
                    <>
                        <button
                            type="button"
                            onClick={toggleQueryForm}
                            className="btn btn-link font-weight-bold p-0 text-left test-trigger-button"
                        >
                            When there are new search results
                        </button>
                        <span className="text-muted">
                            This trigger will fire when new search results are found for a given search query.
                        </span>
                    </>
                )}
                {showQueryForm && (
                    <>
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
                                    Preview results <OpenInNewIcon />
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
                                className="btn btn-outline-secondary mr-1 test-submit-trigger"
                                onClick={editOrCompleteForm}
                                onSubmit={editOrCompleteForm}
                                type="submit"
                                disabled={queryState.kind !== 'VALID'}
                            >
                                Continue
                            </button>
                            <button type="button" className="btn btn-outline-secondary" onClick={editOrCompleteForm}>
                                Cancel
                            </button>
                        </div>
                    </>
                )}
                {triggerCompleted && (
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">When there are new search results</div>
                            <code className="text-muted">{query}</code>
                        </div>
                        <div>
                            <button type="button" onClick={editOrCompleteForm} className="btn btn-link p-0 text-left">
                                Edit
                            </button>
                        </div>
                    </div>
                )}
            </div>
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
