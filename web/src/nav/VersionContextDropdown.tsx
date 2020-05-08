import * as React from 'react'

export const VersionContextDropdown: React.FunctionComponent<{}> = () => (
    <div>
        {window.context.experimentalFeatures?.versionContexts?.map(versionContext => (
            <span>{versionContext.name}</span>
        ))}
    </div>
)
