import React from 'react'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { TreeOrComponentPageResult } from '../../../../graphql-operations'
import { TreeOrComponentHeader } from './TreeOrComponentHeader'

interface Props extends TelemetryProps {
    data: Extract<TreeOrComponentPageResult['node'], { __typename: 'Repository' }>
}

export const TreeOrComponent: React.FunctionComponent<Props> = ({ data, telemetryService, ...props }) => {
    const primaryComponent = data.primaryComponents.length > 0 ? data.primaryComponents[0] : null

    return (
        <>
            <TreeOrComponentHeader primaryComponent={primaryComponent} />
        </>
    )
}
