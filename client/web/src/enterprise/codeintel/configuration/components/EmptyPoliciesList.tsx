import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link } from '@sourcegraph/wildcard'

export const EmptyPoliciesList: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        {'No policies have been defined.  Enable precise code intelligence by '}
        <Link to="/help/code_intelligence/how-to/configure_data_retention" target="_blank" rel="noreferrer noopener">
            configuring data retention policies
        </Link>
        .
    </p>
)
