import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link, LinkProps } from '@sourcegraph/shared/src/components/Link'
import { Button } from '@sourcegraph/wildcard'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<NewBatchChangeButtonProps> = ({ to }) => (
    <Button to={to} variant="primary" as={Link}>
        <PlusIcon className="icon-inline" /> Create batch change
    </Button>
)
