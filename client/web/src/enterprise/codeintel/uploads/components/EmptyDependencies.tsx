import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

export const EmptyDependencies: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        This upload has no dependencies.
    </p>
)
