import H from 'history'
import { sortBy } from 'lodash'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CloseBoxIcon from 'mdi-react/CloseBoxIcon'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { DiagnosticWithType } from '../../../../../../shared/src/api/client/services/diagnosticService'
import { isDefined } from '../../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { Form } from '../../../../components/Form'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticInfo, diagnosticQueryMatcher } from '../../../threadsOLD/detail/backend'
import {
    appendToDiagnosticQuery,
    DiagnosticResolutionStatus,
    isInDiagnosticQuery,
    replaceInDiagnosticQuery,
} from './diagnosticQuery'
import { DiagnosticQueryBuilderQuickFilterDropdownButton } from './DiagnosticQueryBuilderQuickFilterDropdownButton'
import { DiagnosticQueryBuilderFilterDropdownButton } from './DiagnosticQueryBuilderTagFilterDropdownButton'
import { ChangesetPlanProps } from './useChangesetPlan'

interface Props extends Pick<ChangesetPlanProps, 'changesetPlan'>, QueryParameterProps {
    defaultQuery: string
    diagnostics: DiagnosticInfo[]

    className?: string
    location: H.Location
}

/**
 * A query builder for a diagnostic query.
 */
export const DiagnosticQueryBuilder: React.FunctionComponent<Props> = ({
    defaultQuery,
    diagnostics,
    changesetPlan,
    query,
    onQueryChange,
    className = '',
    location,
}) => {
    const [uncommittedValue, setUncommittedValue] = useState(query)
    useEffect(() => setUncommittedValue(query), [query])

    const [isFocused, setIsFocused] = useState(false)
    const onFocus = useCallback(() => setIsFocused(true), [])
    const onBlur = useCallback(() => setIsFocused(false), [])

    const onSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            onQueryChange(uncommittedValue)
        },
        [onQueryChange, uncommittedValue]
    )
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setUncommittedValue(e.currentTarget.value),
        []
    )

    const urlToQuery = useCallback(
        (diagnosticQuery: string): H.LocationDescriptor => {
            const params = new URLSearchParams(location.search)
            params.set('q', diagnosticQuery)
            return `?${params.toString()}`
        },
        [location.search]
    )

    const diagnosticStatuses = useMemo((): Record<DiagnosticResolutionStatus, number> => {
        const counts: Record<DiagnosticResolutionStatus, number> = { unresolved: 0, pending: 0 }
        const matchers = changesetPlan.operations
            .map(op => (op.diagnostics ? diagnosticQueryMatcher(op.diagnostics) : undefined))
            .filter(isDefined)
        const matchesAny = (d: DiagnosticWithType): boolean => matchers.some(m => m(d))
        for (const d of diagnostics) {
            if (matchesAny(d)) {
                counts.pending++
            } else {
                counts.unresolved++
            }
        }
        return counts
    }, [changesetPlan.operations, diagnostics])

    const diagnosticTags = useMemo(() => {
        const counts: { [tag: string]: number } = {}
        for (const tag of diagnostics.flatMap(d => d.tags || [])) {
            counts[tag] = (counts[tag] || 0) + 1
        }
        return sortBy(Object.entries(counts), 1).reverse()
    }, [diagnostics])

    const diagnosticRepositories = useMemo(() => {
        const counts: { [repo: string]: number } = {}
        for (const diagnostic of diagnostics) {
            const repo = parseRepoURI(diagnostic.resource.toString()).repoName
            counts[repo] = (counts[repo] || 0) + 1
        }
        return sortBy(Object.entries(counts), 1).reverse()
    }, [diagnostics])

    return (
        <div className={`diagnostic-query-builder d-flex align-items-center ${className}`}>
            <Link
                to={urlToQuery(replaceInDiagnosticQuery(query, 'is:', DiagnosticResolutionStatus.Unresolved))}
                className={`btn btn-link ${isInDiagnosticQuery(query, 'is:', DiagnosticResolutionStatus.Unresolved)}`}
                onClick={() => alert('TODO: unimplemented')}
            >
                <AlertCircleOutlineIcon className="icon-inline mr-1" />
                {diagnosticStatuses.unresolved} unresolved
            </Link>
            <Link
                to={urlToQuery(replaceInDiagnosticQuery(query, 'is:', DiagnosticResolutionStatus.PendingResolution))}
                className={`btn btn-link ${isInDiagnosticQuery(
                    query,
                    'is:',
                    DiagnosticResolutionStatus.PendingResolution
                )}`}
                onClick={() => alert('TODO: unimplemented')}
            >
                <ProgressCheckIcon className="icon-inline mr-1" />
                {diagnosticStatuses.pending} pending resolution
            </Link>
            <DiagnosticQueryBuilderFilterDropdownButton
                items={diagnosticTags.map(([tag, count]) => ({
                    text: tag,
                    url: urlToQuery(appendToDiagnosticQuery(query, 'tag:', tag)),
                    count,
                }))}
                pluralNoun="tags"
                buttonText="Tags"
                headerText="Filter by tag"
                queryPlaceholderText="Filter tags"
                buttonClassName="btn-link"
            />
            <DiagnosticQueryBuilderFilterDropdownButton
                items={diagnosticRepositories.map(([repo, count]) => ({
                    text: repo,
                    url: urlToQuery(appendToDiagnosticQuery(query, 'repo:', repo)),
                    count,
                }))}
                pluralNoun="repositories"
                buttonText="Repositories"
                headerText="Filter by repository"
                queryPlaceholderText="Filter repositories"
                buttonClassName="btn-link"
            />
            <Form className="flex-1 form d-flex align-items-stretch" onSubmit={onSubmit}>
                <DiagnosticQueryBuilderQuickFilterDropdownButton buttonClassName="px-3 d-none" />
                <div className={`input-group bg-form-control border ${isFocused ? 'form-control-focus' : ''}`}>
                    <div className="input-group-prepend">
                        <span className="input-group-text border-0 pl-2 pr-1 bg-form-control">
                            <SearchIcon className="icon-inline" />
                        </span>
                    </div>
                    <input
                        type="text"
                        className="form-control border-0 pl-1"
                        aria-label="Filter diagnostics"
                        autoCapitalize="off"
                        value={uncommittedValue}
                        onChange={onChange}
                        onFocus={onFocus}
                        onBlur={onBlur}
                    />
                    <div className="input-group-append">
                        <Link to={urlToQuery(defaultQuery)} className="btn btn-link">
                            <CloseBoxIcon className="icon-inline mr-2" />
                            Clear filters
                        </Link>
                    </div>
                </div>
            </Form>
        </div>
    )
}
