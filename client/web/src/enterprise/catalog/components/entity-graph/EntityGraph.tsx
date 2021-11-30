import React from 'react'

import { CatalogGraphFields } from '../../../../graphql-operations'

interface Props {
    graph: CatalogGraphFields
    className?: string
}

export const EntityGraph: React.FunctionComponent<Props> = ({ graph, className }) => (
    <pre>{JSON.stringify(graph, null, 2)}</pre>
)
