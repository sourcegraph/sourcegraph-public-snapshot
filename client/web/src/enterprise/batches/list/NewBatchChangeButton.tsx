import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link, LinkProps, Button, Icon } from '@sourcegraph/wildcard'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<NewBatchChangeButtonProps> = ({ to }) => (
    <Button to={to} variant="primary" as={Link}>
        <Icon as={PlusIcon} /> Create batch change
    </Button>
)
