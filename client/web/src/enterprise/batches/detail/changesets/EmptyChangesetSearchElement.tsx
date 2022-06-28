import React from 'react'

import MagnifyIcon from 'mdi-react/MagnifyIcon'

export const EmptyChangesetSearchElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted row mb-3 w-100">
        <div className="col-12 text-center">
            <MagnifyIcon className="icon" />
            <div className="pt-2">No changesets matched the search and/or filters selected.</div>
        </div>
    </div>
)
