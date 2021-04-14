import ArchiveIcon from 'mdi-react/ArchiveIcon'
import React from 'react'

export const EmptyArchivedChangesetListElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted mt-4 pt-4 mb-4 row">
        <div className="col-12 text-center">
            <ArchiveIcon className="icon" />
            <div className="pt-2">This batch change does not contain archived changesets.</div>
        </div>
    </div>
)
