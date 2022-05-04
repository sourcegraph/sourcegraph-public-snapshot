import React from 'react'

import PlusIcon from 'mdi-react/PlusIcon'

import { Link, LinkProps, Button, Icon } from '@sourcegraph/wildcard'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<React.PropsWithChildren<NewBatchChangeButtonProps>> = ({
    to,
}) => (
    <Button to={to} variant="primary" as={Link}>
        <Icon as={PlusIcon} /> Create batch change
    </Button>
)
