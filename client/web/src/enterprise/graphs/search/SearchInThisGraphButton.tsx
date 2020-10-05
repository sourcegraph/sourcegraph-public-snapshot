import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback } from 'react'
import { GraphSelectionProps, SelectableGraph } from '../selector/graphSelectionProps'

interface Props extends Pick<GraphSelectionProps, 'setSelectedGraph'> {
    graph: SelectableGraph
    query?: string
    className?: string
}

export const SearchInThisGraphButton: React.FunctionComponent<Props> = ({
    graph,
    query,
    setSelectedGraph,
    className = '',
    children,
}) => {
    const onClick = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()

            setSelectedGraph(graph.id)

            // TODO(sqs): hack
            const searchInput = document.querySelector<HTMLInputElement>('.query-input2 input')
            if (searchInput) {
                searchInput.focus()
                searchInput.value = query ? `${query} ` : ''
                searchInput.setSelectionRange(searchInput.value.length, searchInput.value.length)
            }
        },
        [graph.id, query, setSelectedGraph]
    )

    return window.context?.graphsEnabled ? (
        <button type="button" className={`btn btn-outline-secondary ${className}`} onClick={onClick}>
            <SearchIcon className="icon-inline" /> {children}
        </button>
    ) : null
}
