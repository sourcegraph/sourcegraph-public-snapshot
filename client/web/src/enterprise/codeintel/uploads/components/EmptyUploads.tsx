import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link } from '@sourcegraph/wildcard'

export const EmptyUploads: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        No uploads yet. Enable precise code intelligence by{' '}
        <Link
            to="/help/code_intelligence/explanations/precise_code_intelligence"
            target="_blank"
            rel="noreferrer noopener"
        >
            uploading LSIF data
        </Link>
        .
    </p>
)
