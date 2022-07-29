import React from 'react'

import { mdiMapSearch } from '@mdi/js'

import { Link, Text, Icon } from '@sourcegraph/wildcard'

export const EmptyLockfiles: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No lockfiles indexed yet. Enable lockfile indexing by{' '}
        <Link to="/help/code_search/how-to/dependencies_search" target="_blank" rel="noreferrer noopener">
            following the instructions here
        </Link>
        .
    </Text>
)
