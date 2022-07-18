import React from 'react'

import { mdiPlus } from '@mdi/js'

import { Link, LinkProps, Button, Icon } from '@sourcegraph/wildcard'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<React.PropsWithChildren<NewBatchChangeButtonProps>> = ({
    to,
}) => (
    <Button to={to} variant="primary" as={Link}>
        <Icon aria-hidden={true} svgPath={mdiPlus} /> Create batch change
    </Button>
)
