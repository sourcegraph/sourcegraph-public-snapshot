import React from 'react'

import { mdiMapSearch } from '@mdi/js'

import { Link, Text, Icon } from '@sourcegraph/wildcard'

export const EmptyUploads: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No uploads yet. Enable precise code navigation by{' '}
        <Link to="/help/code_navigation/explanations/precise_code_navigation" target="_blank" rel="noreferrer noopener">
            uploading LSIF data
        </Link>
        .
    </Text>
)
