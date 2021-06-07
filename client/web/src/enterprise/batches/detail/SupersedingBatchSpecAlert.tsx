import { parseISO } from 'date-fns'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { DismissibleAlert } from '../../../components/DismissibleAlert'
import { Timestamp } from '../../../components/time/Timestamp'
import { SupersedingBatchSpecFields } from '../../../graphql-operations'

export interface SupersedingBatchSpecAlertProps {
    spec: SupersedingBatchSpecFields | null
}

export const SupersedingBatchSpecAlert: React.FunctionComponent<SupersedingBatchSpecAlertProps> = ({ spec }) => {
    if (!spec) {
        return <></>
    }

    const { applyURL, createdAt } = spec
    return (
        <DismissibleAlert
            className="alert-info"
            partialStorageKey={`superseding-spec-${parseISO(spec.createdAt).getTime()}`}
        >
            <div className="d-flex align-items-center">
                <div className="flex-grow-1">
                    A <Link to={applyURL}>modified batch spec</Link> was uploaded but not applied{' '}
                    <Timestamp date={createdAt} />.
                </div>
            </div>
        </DismissibleAlert>
    )
}
