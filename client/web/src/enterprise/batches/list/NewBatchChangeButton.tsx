import React from 'react'

import PlusIcon from 'mdi-react/PlusIcon'

import { Link, Button, Icon } from '@sourcegraph/wildcard'

export const NewBatchChangeButton: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Button to="/batch-changes/create" variant="primary" as={Link}>
        <Icon as={PlusIcon} /> Create batch change
    </Button>
)
