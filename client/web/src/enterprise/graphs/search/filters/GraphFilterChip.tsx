import React, { useCallback, useMemo } from 'react'
import { FilterChip } from '../../../../search/FilterChip'
import { SelectableGraph } from '../../selector/graphSelectionProps'

interface Props {
    graph: SelectableGraph
    isSelected: boolean
    onSelect: (graph: SelectableGraph) => void
}

export const GraphFilterChip: React.FunctionComponent<Props> = ({ graph, isSelected, onSelect }) => {
    // TODO(sqs): hacky
    const count = useMemo(() => Math.floor(Math.random() * 7 * graph.name.length), [])

    const onFilterChosen = useCallback(() => onSelect(graph), [])
    return (
        <FilterChip
            name={graph.name}
            onFilterChosen={onFilterChosen}
            count={count}
            //
            // hack to show selected graph
            query={isSelected ? ' ' : ''}
            value=" "
        />
    )
}
