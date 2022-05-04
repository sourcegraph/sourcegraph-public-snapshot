import React from 'react'

import ArchiveIcon from 'mdi-react/ArchiveIcon'

export const EmptyArchivedChangesetListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted mb-3 text-center w-100">
        <ArchiveIcon className="icon" />
        <div className="pt-2">This batch change does not contain archived changesets.</div>
    </div>
)
