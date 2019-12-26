import React, { useCallback } from 'react'
import useLocalStorage from 'react-use-localstorage'
import { QueryBuilderProps, QueryBuilder } from './QueryBuilder'

interface Props extends QueryBuilderProps {
    buttonClassName?: string
}

const STORAGE_KEY = 'query-builder-open'

/**
 * A toggle link for expanding and collapsing {@link QueryBuilder}. When expanded, this component also
 * renders the entire query builder. The expanded/collapsed state is persisted in localStorage.
 */
export const QueryBuilderExpandCollapseLink: React.FunctionComponent<Props> = ({ buttonClassName = '', ...props }) => {
    const [isExpandedStr, setIsExpandedStr] = useLocalStorage(STORAGE_KEY, '')
    const isExpanded = isExpandedStr !== ''
    const onToggleClick = useCallback(() => {
        setIsExpandedStr(isExpanded ? '' : 'true')
    }, [isExpanded, setIsExpandedStr])

    return (
        <>
            <button type="button" onClick={onToggleClick} className={`btn btn-link ${buttonClassName}`}>
                {isExpanded ? 'Hide' : 'Use'} search query builder
            </button>
            {isExpanded && <QueryBuilder {...props} />}
        </>
    )
}
