import React from 'react'

import { mdiArchive } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

export const EmptyArchivedChangesetListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted mb-3 text-center w-100">
        <Icon className="icon" svgPath={mdiArchive} inline={false} aria-hidden={true} />
        <div className="pt-2">This batch change does not contain archived changesets.</div>
    </div>
)
