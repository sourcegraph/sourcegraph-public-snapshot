import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

export const EmptyDependencies: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <Icon as={MapSearchIcon} inline={false} className="mb-2" />
        <br />
        This upload has no dependencies.
    </p>
)
