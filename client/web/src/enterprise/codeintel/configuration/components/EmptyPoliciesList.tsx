import React from 'react'

import { mdiMapSearch } from '@mdi/js'

import { Link, Text, Icon } from '@sourcegraph/wildcard'

export const EmptyPoliciesList: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1" data-testid="summary">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        {'No policies have been defined.  Enable precise code navigation by '}
        <Link to="/help/code_navigation/how-to/configure_data_retention" target="_blank" rel="noreferrer noopener">
            configuring data retention policies
        </Link>
        .
    </Text>
)
