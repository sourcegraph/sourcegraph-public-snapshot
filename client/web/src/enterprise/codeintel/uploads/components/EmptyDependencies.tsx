import React from 'react'

import { mdiMapSearch } from '@mdi/js'

import { Text, Icon } from '@sourcegraph/wildcard'

export const EmptyDependencies: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        This upload has no dependencies.
    </Text>
)
