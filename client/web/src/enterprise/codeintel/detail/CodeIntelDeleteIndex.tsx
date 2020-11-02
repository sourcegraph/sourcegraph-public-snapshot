import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { FunctionComponent } from 'react'
import { ErrorLike } from '../../../../../shared/src/util/errors'

export interface CodeIntelDeleteIndexProps {
    deleteIndex: () => Promise<void>
    deletionOrError?: 'loading' | 'deleted' | ErrorLike
}

export const CodeIntelDeleteIndex: FunctionComponent<CodeIntelDeleteIndexProps> = ({
    deleteIndex,
    deletionOrError,
}) => (
    <div className="mt-4 p-2">
        <button
            type="button"
            className="btn btn-danger"
            onClick={deleteIndex}
            disabled={deletionOrError === 'loading'}
            aria-describedby="upload-delete-button-help"
        >
            <DeleteIcon className="icon-inline" /> Delete index
        </button>

        <small id="upload-delete-button-help" className="form-text text-muted">
            Deleting this index will remove it from the index queue.
        </small>
    </div>
)
