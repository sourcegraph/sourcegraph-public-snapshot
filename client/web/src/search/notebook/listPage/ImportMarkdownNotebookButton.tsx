import UploadIcon from 'mdi-react/UploadIcon'
import React, { useCallback, useRef } from 'react'
import * as uuid from 'uuid'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { CreateNotebookVariables } from '../../../graphql-operations'
import { convertMarkdownToBlocks } from '../convertMarkdownToBlocks'
import { blockToGQLInput } from '../serialize'

interface ImportMarkdownNotebookButtonProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
    importNotebook: (notebook: CreateNotebookVariables['notebook']) => void
    isImporting: boolean
    setIsImporting: (value: boolean) => void
}

export const ImportMarkdownNotebookButton: React.FunctionComponent<ImportMarkdownNotebookButtonProps> = ({
    authenticatedUser,
    telemetryService,
    isImporting,
    setIsImporting,
    importNotebook,
}) => {
    const fileInputReference = useRef<HTMLInputElement>(null)

    const onImportButtonClick = useCallback(() => {
        telemetryService.log('SearchNotebookImportMarkdownNotebookButtonClick')
        fileInputReference.current?.click()
    }, [fileInputReference, telemetryService])

    const onFileInputChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const files = event.target.files
            if (!files || files.length !== 1) {
                return
            }
            setIsImporting(true)
            const fileName = files[0].name
            // TODO: Check file size
            const reader = new FileReader()
            reader.addEventListener('load', event => {
                if (!event.target || !event.target.result || typeof event.target.result !== 'string') {
                    return
                }
                const blocks = convertMarkdownToBlocks(event.target.result).map(block =>
                    blockToGQLInput({
                        id: uuid.v4(),
                        ...block,
                    })
                )
                importNotebook({
                    title: fileName.split('.snb.md')[0],
                    blocks,
                    public: false,
                    namespace: authenticatedUser.id,
                })
            })
            reader.readAsText(files[0])
        },
        [authenticatedUser, importNotebook, setIsImporting]
    )

    return (
        <>
            <input type="file" className="d-none" ref={fileInputReference} accept=".md" onChange={onFileInputChange} />
            <Button variant="secondary" onClick={onImportButtonClick} disabled={isImporting}>
                <UploadIcon className="icon-inline mr-1" />
                <span>{isImporting ? 'Importing...' : 'Import Markdown notebook'}</span>
            </Button>
        </>
    )
}
