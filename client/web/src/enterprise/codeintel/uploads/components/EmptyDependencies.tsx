import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'

export const EmptyDependencies: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        This upload has no dependencies.
    </p>
)
