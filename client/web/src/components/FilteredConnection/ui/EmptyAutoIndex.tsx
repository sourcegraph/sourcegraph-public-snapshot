import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'

interface EmptyAutoIndexProps {
    pluralNoun: string
}
export const EmptyAutoIndex: React.FunctionComponent<EmptyAutoIndexProps> = ({ pluralNoun }) => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        {`No ${pluralNoun} yet.  Enable precise code intelligence by `}
        <a
            href="https://docs.sourcegraph.com/code_intelligence/how-to/index_a_go_repository"
            target="_blank"
            rel="noreferrer noopener"
        >
            auto-indexing LSIF data
        </a>
        .
    </p>
)
