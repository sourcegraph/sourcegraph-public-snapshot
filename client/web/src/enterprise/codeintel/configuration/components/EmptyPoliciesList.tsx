import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'

export const EmptyPoliciesList: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        {'No policies have been defined.  Enable precise code intelligence by '}
        <a
            href="https://docs.sourcegraph.com/code_intelligence/how-to/configure_data_retention"
            target="_blank"
            rel="noreferrer noopener"
        >
            configuring data retention policies
        </a>
        .
    </p>
)
