import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link, Button } from '@sourcegraph/wildcard'
import type { LinkProps } from '@sourcegraph/wildcard/src/components/Link'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<NewBatchChangeButtonProps> = ({ to }) => (
    <Button to={to} variant="primary" as={Link}>
        <PlusIcon className="icon-inline" /> Create batch change
    </Button>
)
