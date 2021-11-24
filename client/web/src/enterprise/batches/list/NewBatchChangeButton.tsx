import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link, LinkProps } from '@sourcegraph/shared/src/components/Link'

interface NewBatchChangeButtonProps extends Pick<LinkProps, 'to'> {}

export const NewBatchChangeButton: React.FunctionComponent<NewBatchChangeButtonProps> = ({ to }) => (
    <Link to={to} className="btn btn-primary">
        <PlusIcon className="icon-inline" /> Create Batch Change
    </Link>
)
