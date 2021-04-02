import * as H from 'history'
import React, { useEffect, useState } from 'react'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Link } from '@sourcegraph/shared/src/components/Link'

export interface ChangesetsArchivedNoticeProps {
    history: H.History
    location: H.Location
}

export const ChangesetsArchivedNotice: React.FunctionComponent<ChangesetsArchivedNoticeProps> = ({
    history,
    location,
}) => {
    const [archivedCount, setArchivedCount] = useState<number | undefined>()
    const [archivedBy, setArchivedBy] = useState<string | undefined>()
    useEffect(() => {
        const parameters = new URLSearchParams(location.search)

        const count = parameters.get('archivedCount')
        parameters.delete('archivedCount')
        const archived = parameters.get('archivedBy')
        parameters.delete('archivedBy')
        if (count !== null && archived !== null) {
            setArchivedCount(parseInt(count, 10))
            setArchivedBy(archived)
        }

        if (new URLSearchParams(location.search).toString() !== parameters.toString()) {
            history.replace({ ...location, search: parameters.toString() })
        }
    }, [history, location])

    if (!archivedCount || !archivedBy) {
        return <></>
    }

    return (
        <DismissibleAlert className="alert alert-info" partialStorageKey={`changesets-archived-by-${archivedBy}`}>
            <div className="d-flex align-items-center">
                <div className="d-none d-md-block">
                    <ArchiveIcon className="icon icon-inline mr-2" />
                </div>
                <div className="flex-grow-1">
                    {archivedCount} {pluralize('changeset', archivedCount)} {pluralize('has', archivedCount, 'have')}{' '}
                    been <Link to="?tab=archived">archived</Link>.
                </div>
            </div>
        </DismissibleAlert>
    )
}
