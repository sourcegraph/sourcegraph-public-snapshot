import React from 'react'

import { parseISO } from 'date-fns'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../components/DismissibleAlert'
import type { SupersedingBatchSpecFields } from '../../../graphql-operations'

export interface SupersedingBatchSpecAlertProps {
    spec: SupersedingBatchSpecFields | null
}

export const SupersedingBatchSpecAlert: React.FunctionComponent<
    React.PropsWithChildren<SupersedingBatchSpecAlertProps>
> = ({ spec }) => {
    if (!spec) {
        return <></>
    }

    const { applyURL, createdAt } = spec

    if (applyURL === null) {
        return null
    }

    return (
        <DismissibleAlert variant="info" partialStorageKey={`superseding-spec-${parseISO(spec.createdAt).getTime()}`}>
            <div className="d-flex align-items-center">
                <div className="flex-grow-1">
                    A <Link to={applyURL}>modified batch spec</Link> was uploaded{' '}
                    <Timestamp date={createdAt} noAbout={true} />, but has not been applied.
                </div>
            </div>
        </DismissibleAlert>
    )
}
