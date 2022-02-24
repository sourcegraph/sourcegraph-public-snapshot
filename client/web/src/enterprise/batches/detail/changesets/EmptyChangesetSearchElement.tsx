import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

export const EmptyChangesetSearchElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted row mb-3 w-100">
        <div className="col-12 text-center">
            <Icon inline={false} as={MagnifyIcon} className="icon" />
            <div className="pt-2">No changesets matched the search and/or filters selected.</div>
        </div>
    </div>
)
