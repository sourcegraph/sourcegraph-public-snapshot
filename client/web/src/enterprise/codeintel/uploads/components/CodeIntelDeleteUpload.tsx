import { FunctionComponent } from 'react'

import DeleteIcon from 'mdi-react/DeleteIcon'

import { ErrorLike } from '@sourcegraph/common'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Icon } from '@sourcegraph/wildcard'

export interface CodeIntelDeleteUploadProps {
    state: LSIFUploadState
    deleteUpload: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

export const CodeIntelDeleteUpload: FunctionComponent<React.PropsWithChildren<CodeIntelDeleteUploadProps>> = ({
    state,
    deleteUpload,
    deletionOrError,
}) =>
    state === LSIFUploadState.DELETING ? (
        <></>
    ) : (
        <Button
            type="button"
            variant="danger"
            onClick={deleteUpload}
            disabled={deletionOrError === 'loading'}
            aria-describedby="upload-delete-button-help"
            data-tooltip={
                state === LSIFUploadState.COMPLETED
                    ? 'Deleting this upload will make it unavailable to answer code intelligence queries the next time the repository commit graph is refreshed.'
                    : 'Delete this upload immediately'
            }
        >
            <Icon as={DeleteIcon} /> Delete upload
        </Button>
    )
