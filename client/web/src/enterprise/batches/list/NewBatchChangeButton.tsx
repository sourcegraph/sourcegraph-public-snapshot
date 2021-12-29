import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { RouterLink } from '@sourcegraph/wildcard'
import type { LinkProps } from '@sourcegraph/wildcard/src/components/Link'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<NewBatchChangeButtonProps> = ({ to }) => (
    <RouterLink to={to} className="btn btn-primary">
        <PlusIcon className="icon-inline" /> Create batch change
    </RouterLink>
)
