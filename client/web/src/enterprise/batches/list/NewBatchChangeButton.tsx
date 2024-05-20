import React from 'react'

import { mdiPlus } from '@mdi/js'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Link, type LinkProps, Button, Icon, Tooltip } from '@sourcegraph/wildcard'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'>, TelemetryV2Props {
    // canCreate indicates whether or not the currently-authenticated user has sufficient
    // permissions to create a batch change in whatever context this button is being
    // presented. If not, canCreate should be a string reason why the user cannot create
    // to be used for the button tooltip.
    canCreate: true | string
}

export const NewBatchChangeButton: React.FunctionComponent<React.PropsWithChildren<NewBatchChangeButtonProps>> = ({
    canCreate,
    to,
    telemetryRecorder,
}) => {
    const button = (
        <Button
            disabled={typeof canCreate === 'string'}
            to={to}
            variant="primary"
            as={Link}
            onClick={() => {
                EVENT_LOGGER.log('batch_change_list_page:create_batch_change_details:clicked')
                telemetryRecorder.recordEvent('batchChanges', 'create')
            }}
        >
            <Icon aria-hidden={true} svgPath={mdiPlus} /> Create batch change
        </Button>
    )
    return typeof canCreate === 'string' ? <Tooltip content={canCreate}>{button}</Tooltip> : button
}
