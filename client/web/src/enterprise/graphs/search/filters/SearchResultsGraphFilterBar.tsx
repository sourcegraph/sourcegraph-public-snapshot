import React, { useCallback } from 'react'
import { FilterChip } from '../../../../search/FilterChip'
import { GraphSelectionProps, SelectableGraph } from '../../selector/graphSelectionProps'
import { useGraphs } from '../../selector/useGraphs'
import { GraphFilterChip } from './GraphFilterChip'

interface Props extends GraphSelectionProps {
    className?: string
    listClassName?: string
    'data-testid'?: string
}

const NULL_GRAPH_ID = 'null'

export const SearchResultsGraphFilterBar: React.FunctionComponent<Props> = ({
    selectedGraph,
    setSelectedGraph,
    className = '',
    listClassName = '',
    'data-testid': dataTestId,
    ...props
}) => {
    const graphs = useGraphs(props)

    const onGraphSelect = useCallback((graph: SelectableGraph): void => setSelectedGraph(graph.id), [])

    return (
        <div className={className} data-testid={dataTestId}>
            Graphs:
            <div className={listClassName}>
                {graphs.map(graph => (
                    <GraphFilterChip
                        key={graph.id === null ? NULL_GRAPH_ID : graph.id}
                        graph={graph}
                        isSelected={selectedGraph === graph.id}
                        onSelect={onGraphSelect}
                    />
                ))}
            </div>
        </div>
    )
}
