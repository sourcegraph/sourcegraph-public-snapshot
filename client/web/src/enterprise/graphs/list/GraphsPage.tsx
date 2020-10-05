import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { GraphSelectionProps } from '../selector/graphSelectionProps'
import { useGraphs } from '../selector/useGraphs'
import { GraphList } from '../shared/graphList/GraphList'

interface Props extends GraphSelectionProps {}

export const GraphsPage: React.FunctionComponent<Props> = ({ ...props }) => {
    const graphs = useGraphs(props)

    return (
        <div className="container mt-3">
            <div className="d-flex align-items-center justify-content-between mb-2">
                <h1>Graphs</h1>
                {/* TODO(sqs): support adding graphs in their orgs too */}
                <Link to="/user/graphs/new" className="btn btn-secondary">
                    <PlusIcon className="icon-inline" /> New graph
                </Link>
            </div>

            <GraphList {...props} graphs={{ nodes: graphs }} />
        </div>
    )
}
