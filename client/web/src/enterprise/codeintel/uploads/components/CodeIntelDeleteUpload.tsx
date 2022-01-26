import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { FunctionComponent } from 'react'

import { ErrorLike } from '@sourcegraph/common'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { Button } from '@sourcegraph/wildcard'

export interface CodeIntelDeleteUploadProps {
    state: LSIFUploadState
    deleteUpload: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

export const CodeIntelDeleteUpload: FunctionComponent<CodeIntelDeleteUploadProps> = ({
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
            <DeleteIcon className="icon-inline" /> Delete upload
        </Button>
    )
