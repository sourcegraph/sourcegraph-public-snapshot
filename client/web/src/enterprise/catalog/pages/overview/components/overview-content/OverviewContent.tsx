import React, { useContext, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { authenticatedUser } from '@sourcegraph/web/src/auth'

import { CatalogBackendContext } from '../../../../core/backend/context'

export interface OverviewContentProps extends TelemetryProps {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<OverviewContentProps> = props => {
    const { telemetryService } = props

    const history = useHistory()
    const { listComponents } = useContext(CatalogBackendContext)

    const components = useObservable(useMemo(() => listComponents(), [listComponents]))

    const user = useObservable(authenticatedUser)

    if (components === undefined) {
        return <LoadingSpinner />
    }

    return (
        <div>
            <section className="d-flex flex-wrap align-items-center">
                Foos: <code>{JSON.stringify(components)}</code>
                <br />
                User: {user?.username || 'none'}
            </section>
        </div>
    )
}
