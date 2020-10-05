import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'
import { GraphListItem as GraphListItemFragment } from '../../../../graphql-operations'
import { GraphIcon } from '../../icons'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { GraphSelectionProps } from '../../selector/graphSelectionProps'
import CheckIcon from 'mdi-react/CheckIcon'

export const GraphListItemFragmentGQL = gql`
    fragment GraphListItem on Graph {
        id
        name
        description
        url
        editURL
    }
`

type WithOptionalKeys<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>

/**
 * Props that should be passed to {@link GraphList}.
 */
export interface GraphListItemProps extends Pick<GraphSelectionProps, 'selectedGraph' | 'setSelectedGraph'> {}

interface Props extends GraphListItemProps {
    node: WithOptionalKeys<Omit<GraphListItemFragment, 'id'>, 'url' | 'editURL'> & {
        id: GraphListItemFragment['id'] | null
    }
    className?: string
}

export const GraphListItem: React.FunctionComponent<Props> = ({
    node: graph,
    selectedGraph,
    setSelectedGraph,
    className = '',
}) => {
    const onSelect = useCallback(() => setSelectedGraph(graph.id), [setSelectedGraph])
    const isSelected = graph.id === selectedGraph
    return (
        <li className={`d-flex align-items-start ${className}`}>
            <GraphIcon className="mt-1 mr-2 icon-inline text-muted" />
            <header className="flex-1 mr-3">
                <h3 className="mb-0">
                    <LinkOrSpan to={graph.url}>{graph.name}</LinkOrSpan>
                </h3>
                {graph.description && <small className="text-muted">{graph.description}</small>}
            </header>
            {graph.editURL && (
                <Link to={graph.editURL} className="btn btn-secondary btn-sm mr-1">
                    Edit
                </Link>
            )}
            {isSelected ? (
                <button type="button" disabled={true} className="btn btn-success btn-sm">
                    <CheckIcon className="icon-inline" /> Selected
                </button>
            ) : (
                <button type="button" className="btn btn-secondary btn-sm" onClick={onSelect}>
                    Select
                </button>
            )}
        </li>
    )
}
