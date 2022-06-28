import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link, Text } from '@sourcegraph/wildcard'

export const EmptyAutoIndex: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        {'No indexes yet.  Enable precise code intelligence by '}
        <Link to="/help/code_intelligence/how-to/enable_auto_indexing" target="_blank" rel="noreferrer noopener">
            auto-indexing LSIF data
        </Link>
        .
    </Text>
)
