import { FunctionComponent } from 'react'

import DeleteIcon from 'mdi-react/DeleteIcon'

import { ErrorLike } from '@sourcegraph/common'
import { Button, Icon } from '@sourcegraph/wildcard'

export interface CodeIntelDeleteIndexProps {
    deleteIndex: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

export const CodeIntelDeleteIndex: FunctionComponent<React.PropsWithChildren<CodeIntelDeleteIndexProps>> = ({
    deleteIndex,
    deletionOrError,
}) => (
    <Button
        type="button"
        variant="danger"
        onClick={deleteIndex}
        disabled={deletionOrError === 'loading'}
        aria-describedby="upload-delete-button-help"
        data-tooltip="Deleting this index will remove it from the index queue."
    >
        <Icon as={DeleteIcon} /> Delete index
    </Button>
)
