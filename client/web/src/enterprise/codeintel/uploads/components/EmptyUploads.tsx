import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Link, Text } from '@sourcegraph/wildcard'

export const EmptyUploads: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
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
    </Text>
)
