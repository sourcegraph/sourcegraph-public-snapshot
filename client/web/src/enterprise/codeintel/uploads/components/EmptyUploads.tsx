import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'

export const EmptyUploads: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        No uploads yet. Enable precise code intelligence by{' '}
        <a
            href="https://docs.sourcegraph.com/code_intelligence/explanations/precise_code_intelligence"
            target="_blank"
            rel="noreferrer noopener"
        >
            uploading LSIF data
        </a>
        .
    </p>
)
