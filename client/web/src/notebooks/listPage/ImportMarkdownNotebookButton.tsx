import React, { useCallback, useRef } from 'react'

import UploadIcon from 'mdi-react/UploadIcon'
import * as uuid from 'uuid'

import { ErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CreateNotebookVariables } from '../../graphql-operations'
import { blockToGQLInput } from '../serialize'
import { convertMarkdownToBlocks } from '../serialize/convertMarkdownToBlocks'

const LOADING = 'loading' as const

const INVALID_IMPORT_FILE_ERROR = new Error(
    'Cannot read the imported file. Check that the imported file is a Markdown-formatted text file.'
)

const MAX_FILE_SIZE_IN_BYTES = 1000 * 1000 // 1MB

interface ImportMarkdownNotebookButtonProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
    importNotebook: (notebook: CreateNotebookVariables['notebook']) => void
    importState: typeof LOADING | ErrorLike | undefined
    setImportState: (state: typeof LOADING | ErrorLike | undefined) => void
}

export const ImportMarkdownNotebookButton: React.FunctionComponent<
    React.PropsWithChildren<ImportMarkdownNotebookButtonProps>
> = ({ authenticatedUser, telemetryService, importState, setImportState, importNotebook }) => {
    const fileInputReference = useRef<HTMLInputElement>(null)

    const onImportButtonClick = useCallback(() => {
        telemetryService.log('SearchNotebookImportMarkdownNotebookButtonClick')
        // Open the system file picker.
        fileInputReference.current?.click()
    }, [fileInputReference, telemetryService])

    const onFileLoad = useCallback(
        (event: ProgressEvent<FileReader>, fileName: string): void => {
            if (!event.target || !event.target.result || typeof event.target.result !== 'string') {
                setImportState(INVALID_IMPORT_FILE_ERROR)
                return
            }
            const blocks = convertMarkdownToBlocks(event.target.result).map(block =>
                blockToGQLInput({ id: uuid.v4(), ...block })
            )
            const title = fileName.split('.snb.md')[0].trim() || 'New Notebook'
            importNotebook({ title, blocks, public: false, namespace: authenticatedUser.id })
        },
        [authenticatedUser, importNotebook, setImportState]
    )

    const onFileInputChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const files = event.target.files
            if (!files || files.length !== 1) {
                setImportState(INVALID_IMPORT_FILE_ERROR)
                return
            }

            if (files[0].size > MAX_FILE_SIZE_IN_BYTES) {
                setImportState(new Error('File too large. Maximum allowed file size is 1MB.'))
                return
            }

            setImportState(LOADING)

            const reader = new FileReader()
            reader.addEventListener('load', event => onFileLoad(event, files[0].name))
            reader.readAsText(files[0])
        },
        [setImportState, onFileLoad]
    )

    return (
        <>
            <input
                type="file"
                className="d-none"
                ref={fileInputReference}
                accept=".md"
                onChange={onFileInputChange}
                data-testid="import-markdown-notebook-file-input"
            />
            <Button
                variant="secondary"
                onClick={onImportButtonClick}
                disabled={importState === LOADING}
                className="ml-2"
            >
                <Icon className="mr-1" as={UploadIcon} />
                <span>{importState === LOADING ? 'Importing...' : 'Import Markdown notebook'}</span>
            </Button>
        </>
    )
}
