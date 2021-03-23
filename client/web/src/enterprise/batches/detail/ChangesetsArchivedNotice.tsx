import React from 'react'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import ArchiveIcon from 'mdi-react/ArchiveIcon'

export interface ChangesetsArchivedNoticeProps {
    archivedCount: number
    specID: string
}

export const ChangesetsArchivedNotice: React.FunctionComponent<ChangesetsArchivedNoticeProps> = ({
    archivedCount,
    specID,
}) => {
    if (archivedCount === 0 || specID === '') {
        return <></>
    }

    return (
        <DismissibleAlert className="alert alert-info" partialStorageKey={`changesets-archived-by-${specID}`}>
            <div className="d-flex align-items-center">
                <div className="d-none d-md-block">
                    <ArchiveIcon className="icon icon-inline mr-2" />
                </div>
                <div className="flex-grow-1">
                    {archivedCount === 1
                        ? '1 changeset has been archived'
                        : `${archivedCount} changesets have been archived.`}
                </div>
            </div>
        </DismissibleAlert>
    )
}
