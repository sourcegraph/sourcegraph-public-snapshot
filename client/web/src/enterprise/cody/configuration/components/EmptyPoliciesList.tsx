import React from 'react'

import { mdiMapSearch } from '@mdi/js'

import { H4, Icon } from '@sourcegraph/wildcard'

interface EmptyPoliciesListProps {
    repo?: { id: string; name: string }
}

export const EmptyPoliciesList: React.FunctionComponent<EmptyPoliciesListProps> = ({ repo }) => (
    <div className="d-flex align-items-center flex-column w-100 mt-3" data-testid="summary">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <H4 className="mb-0">No embeddings policies found.</H4>
    </div>
)
