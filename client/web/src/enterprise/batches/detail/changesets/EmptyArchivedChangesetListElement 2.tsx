import ArchiveIcon from 'mdi-react/ArchiveIcon'
import React from 'react'

export const EmptyArchivedChangesetListElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted mb-3 text-center w-100">
        <ArchiveIcon className="icon" />
        <div className="pt-2">This batch change does not contain archived changesets.</div>
    </div>
)
