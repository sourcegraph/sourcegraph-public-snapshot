import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import React, { useEffect, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

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
        <DismissibleAlert className="alert-info" partialStorageKey={`changesets-archived-by-${archivedBy}`}>
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
