import { parseISO } from 'date-fns'
import CreationIcon from 'mdi-react/CreationIcon'
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
                <div className="d-none d-md-block redesign-d-none">
                    <CreationIcon className="icon icon-inline mr-2" />
                </div>
                <div className="flex-grow-1">
                    A <Link to={applyURL}>modified batch spec</Link> was uploaded but not applied{' '}
                    <Timestamp date={createdAt} />.
                </div>
            </div>
        </DismissibleAlert>
    )
}
