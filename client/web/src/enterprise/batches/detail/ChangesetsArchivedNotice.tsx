import React, { useEffect, useState } from 'react'

import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'

import { pluralize } from '@sourcegraph/common'
import { Link, Icon } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'

export interface ChangesetsArchivedNoticeProps {
    history: H.History
    location: H.Location
}

export const ChangesetsArchivedNotice: React.FunctionComponent<
    React.PropsWithChildren<ChangesetsArchivedNoticeProps>
> = ({ history, location }) => {
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
        <DismissibleAlert variant="info" partialStorageKey={`changesets-archived-by-${archivedBy}`}>
            <div className="d-flex align-items-center">
                <div className="d-none d-md-block">
                    <Icon className="icon mr-2" as={ArchiveIcon} />
                </div>
                <div className="flex-grow-1">
                    {archivedCount} {pluralize('changeset', archivedCount)} {pluralize('has', archivedCount, 'have')}{' '}
                    been <Link to="?tab=archived">archived</Link>.
                </div>
            </div>
        </DismissibleAlert>
    )
}
