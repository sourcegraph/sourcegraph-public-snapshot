import React from 'react'
import { Link } from 'react-router-dom'
import { gql } from '../../../../../shared/src/graphql/graphql'
import { GraphTitle as GraphTitleFragment } from '../../../graphql-operations'
import { GraphIcon } from '../icons'

export const GraphTitleGQLFragment = gql`
    fragment GraphTitle on Graph {
        name
        url
        owner {
            ... on Namespace {
                namespaceName
                url
            }
        }
    }
`

interface Props {
    graph: GraphTitleFragment
    tag?: 'h1'
    className?: string
}

export const GraphTitle: React.FunctionComponent<Props> = ({ graph, tag: Tag = 'h1', className = '' }) => (
    <Tag className={className}>
        <GraphIcon />{' '}
        {graph.owner ? <Link to={`${graph.owner.url}/graphs`}>{graph.owner.namespaceName}</Link> : '(deleted)'} /{' '}
        <Link to={graph.url}>{graph.name}</Link>
    </Tag>
)
