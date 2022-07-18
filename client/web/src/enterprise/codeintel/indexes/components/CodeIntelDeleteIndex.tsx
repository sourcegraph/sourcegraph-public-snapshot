import { FunctionComponent } from 'react'

import { mdiDelete } from '@mdi/js'

import { ErrorLike } from '@sourcegraph/common'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

export interface CodeIntelDeleteIndexProps {
    deleteIndex: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

export const CodeIntelDeleteIndex: FunctionComponent<React.PropsWithChildren<CodeIntelDeleteIndexProps>> = ({
    deleteIndex,
    deletionOrError,
}) => (
    <Tooltip content="Deleting this index will remove it from the index queue.">
        <Button type="button" variant="danger" onClick={deleteIndex} disabled={deletionOrError === 'loading'}>
            <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete index
        </Button>
    </Tooltip>
)
