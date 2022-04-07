import React, { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/wildcard'

import { CodeIntelIndexerFields } from '../../../../graphql-operations'

export const CodeIntelIndexer: FunctionComponent<{ indexer: CodeIntelIndexerFields }> = ({ indexer }) =>
    indexer.url === '' ? <>{indexer.name}</> : <Link to={indexer.url}>{indexer.name}</Link>
