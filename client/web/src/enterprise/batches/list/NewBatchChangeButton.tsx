import React from 'react'

import { mdiPlus } from '@mdi/js'

import { Link, LinkProps, Button, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<React.PropsWithChildren<NewBatchChangeButtonProps>> = ({
    to,
}) => (
    <Button
        to={to}
        variant="primary"
        as={Link}
        onClick={() => {
            eventLogger.log('batch_change_list_page:create_batch_change_details:clicked')
        }}
    >
        <Icon aria-hidden={true} svgPath={mdiPlus} /> Create batch change
    </Button>
)
