import {
    ListboxButton,
    ListboxGroupLabel,
    ListboxInput,
    ListboxList,
    ListboxOption,
    ListboxPopover,
} from '@reach/listbox'
import VisuallyHidden from '@reach/visually-hidden'
import { uniqueId } from 'lodash'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React, { useCallback, useMemo } from 'react'
import { cold } from 'react-hot-loader'
import { Link } from 'react-router-dom'
import { SourcegraphContext } from '../../../jscontext'
import { GraphSelectionProps } from './graphSelectionProps'
import { useGraphs } from './useGraphs'

interface Props
    extends Omit<GraphSelectionProps, 'contributeContextualGraphs'>,
        Partial<Pick<SourcegraphContext, 'graphsEnabled'>> {}

const NULL_GRAPH_ID = 'null' // sentinel value because <ListboxOption value> must be non-null

export const GraphSelector: React.FunctionComponent<Props> =
    // Wrap in cold(...) to work around https://github.com/reach/reach-ui/issues/629.
    cold(
        ({
            selectedGraph,
            setSelectedGraph,
            reloadGraphsSeq,
            contextualGraphs,

            // If this uses an optional chain, there is an error `_window$context is not defined`.
            //
            // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
            graphsEnabled = window.context && window.context.graphsEnabled,
        }) => {
            const graphs = useGraphs({ reloadGraphsSeq, contextualGraphs })

            const onChange = useCallback(
                (graphID: string) => {
                    setSelectedGraph(graphID === NULL_GRAPH_ID ? null : graphID)
                },
                [setSelectedGraph]
            )

            const selectedGraphValue = selectedGraph === null ? NULL_GRAPH_ID : selectedGraph

            const labelId = `GraphSelector--${useMemo(() => uniqueId(), [])}`
            return graphsEnabled ? (
                <>
                    <VisuallyHidden id={labelId}>Select graph</VisuallyHidden>
                    <ListboxInput value={selectedGraphValue} onChange={onChange} aria-labelledby={labelId}>
                        <ListboxButton
                            className="btn btn-outline-secondary border-right-0 btn-sm d-inline-flex text-nowrap h-100"
                            arrow={true}
                            style={{ backgroundColor: 'var(--body-bg)' }}
                        >
                            {graphs?.find(graph => graph.id === selectedGraph)?.name || (
                                <DotsHorizontalIcon className="icon-inline" />
                            )}
                        </ListboxButton>
                        <ListboxPopover style={{ maxWidth: '10rem' }}>
                            <ListboxList>
                                {graphs === undefined ? (
                                    <ListboxGroupLabel>Loading...</ListboxGroupLabel>
                                ) : (
                                    <>
                                        {graphs.map(graph => (
                                            <ListboxOption
                                                key={graph.id === null ? NULL_GRAPH_ID : graph.id}
                                                value={graph.id === null ? NULL_GRAPH_ID : graph.id}
                                                title={graph.description}
                                            >
                                                {graph.name}
                                            </ListboxOption>
                                        ))}
                                        <ListboxGroupLabel
                                            className="border-top small mt-2 pt-2"
                                            style={{ whiteSpace: 'unset', minWidth: '10rem' }}
                                        >
                                            <p className="text-muted mb-0">
                                                A graph defines the scope of search and code intelligence.
                                            </p>
                                            <Link className="btn btn-secondary btn-sm mt-1" to="/graphs">
                                                Manage graphs
                                            </Link>
                                        </ListboxGroupLabel>
                                    </>
                                )}
                            </ListboxList>
                        </ListboxPopover>
                    </ListboxInput>
                </>
            ) : null
        }
    )
