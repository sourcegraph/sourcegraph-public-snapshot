import React from 'react'

import { mdiMagnify } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

export const EmptyChangesetSearchElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted row mb-3 w-100">
        <div className="col-12 text-center">
            <Icon className="icon" svgPath={mdiMagnify} inline={false} aria-hidden={true} />
            <div className="pt-2">No changesets matched the search and/or filters selected.</div>
        </div>
    </div>
)
