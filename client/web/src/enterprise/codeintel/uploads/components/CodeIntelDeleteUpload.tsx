import { FunctionComponent } from 'react'

import { mdiDelete } from '@mdi/js'

import { ErrorLike } from '@sourcegraph/common'
import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

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
        <Tooltip
            content={
                state === LSIFUploadState.COMPLETED
                    ? 'Deleting this upload will make it unavailable to answer code navigation queries the next time the repository commit graph is refreshed.'
                    : 'Delete this upload immediately'
            }
        >
            <Button type="button" variant="danger" onClick={deleteUpload} disabled={deletionOrError === 'loading'}>
                <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete upload
            </Button>
        </Tooltip>
    )
