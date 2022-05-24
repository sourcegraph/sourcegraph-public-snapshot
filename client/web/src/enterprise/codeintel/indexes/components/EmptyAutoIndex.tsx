import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link } from '@sourcegraph/wildcard'

export const EmptyAutoIndex: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        {'No indexes yet.  Enable precise code intelligence by '}
        <Link to="/help/code_intelligence/how-to/enable_auto_indexing" target="_blank" rel="noreferrer noopener">
            auto-indexing LSIF data
        </Link>
        .
    </p>
)
