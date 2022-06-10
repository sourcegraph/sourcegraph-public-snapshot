import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Text } from '@sourcegraph/wildcard'

export const EmptyDependencies: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        This upload has no dependencies.
    </Text>
)
