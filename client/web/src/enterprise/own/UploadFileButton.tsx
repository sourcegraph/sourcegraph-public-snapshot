import { useCallback, useRef, useState } from 'react'

import { mdiUpload } from '@mdi/js'

import { type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { ErrorAlert, Icon, Input } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type {
    AddIngestedCodeownersResult,
    AddIngestedCodeownersVariables,
    IngestedCodeowners,
    RepositoryFields,
    UpdateIngestedCodeownersResult,
    UpdateIngestedCodeownersVariables,
} from '../../graphql-operations'

import { ADD_INGESTED_CODEOWNERS_MUTATION, UPDATE_INGESTED_CODEOWNERS_MUTATION } from './graphqlQueries'

// Arbitrary limit to prevent users from uploading huge files.
// We can up it later if we find that perf doesn't suffer and users need it.
const MAX_FILE_SIZE_IN_BYTES = 10 * 1024 * 1024 // 10MB

interface UploadFileButtonProps {
    onComplete: (file: IngestedCodeowners) => void
    repo: RepositoryFields
    fileAlreadyExists: boolean
}

export const UploadFileButton: React.FunctionComponent<UploadFileButtonProps> = ({
    onComplete,
    repo,
    fileAlreadyExists,
}) => {
    const [uploadError, setUploadError] = useState<ErrorLike | null>(null)
    const addCodeonwersFileMutation = useMutation<AddIngestedCodeownersResult, AddIngestedCodeownersVariables>(
        ADD_INGESTED_CODEOWNERS_MUTATION
    )
    const updateCodeownersFileMutation = useMutation<UpdateIngestedCodeownersResult, UpdateIngestedCodeownersVariables>(
        UPDATE_INGESTED_CODEOWNERS_MUTATION
    )

    const [upload, { loading }] = fileAlreadyExists ? updateCodeownersFileMutation : addCodeonwersFileMutation

    const fileInputRef = useRef<HTMLInputElement | null>(null)
    const onUploadClicked = useCallback(() => {
        // Open the system file picker.
        fileInputRef.current?.click()
    }, [])
    const onFileSelected = useCallback(() => {
        const file = fileInputRef.current?.files?.[0]
        if (!file) {
            return
        }

        if (file.size > MAX_FILE_SIZE_IN_BYTES) {
            setUploadError(new Error(`File size must be less than ${MAX_FILE_SIZE_IN_BYTES / 1024 / 1024}MB`))
            return
        }

        const reader = new FileReader()
        reader.addEventListener('load', () => {
            const contents = reader.result as string

            upload({
                variables: {
                    repoID: repo.id,
                    contents,
                },
            })
                .then(result => {
                    if (result.data) {
                        if ('updateCodeownersFile' in result.data) {
                            onComplete(result.data.updateCodeownersFile)
                        } else {
                            onComplete(result.data.addCodeownersFile)
                        }
                    } else {
                        setUploadError(new Error('No data returned from server'))
                    }
                })
                .catch(error => {
                    if (isErrorLike(error)) {
                        setUploadError(error)
                    } else {
                        setUploadError(new Error('Unknown error'))
                    }
                })
                .finally(() => {
                    // Reset the file input so the same file can be reuploaded later.
                    if (fileInputRef.current) {
                        fileInputRef.current.value = ''
                    }
                })
        })
        reader.readAsText(file)
    }, [onComplete, repo.id, upload])

    return (
        <>
            <LoaderButton
                icon={<Icon svgPath={mdiUpload} aria-hidden={true} />}
                label={fileAlreadyExists ? 'Replace current file' : 'Upload file'}
                loading={loading}
                variant="primary"
                onClick={onUploadClicked}
            />
            {uploadError && <ErrorAlert className="mt-2" error={uploadError} prefix="Error uploading file:" />}
            {/* Don't show the file input, the nicer-looking button will trigger it programmatically */}
            <Input ref={fileInputRef} type="file" className="d-none" onChange={onFileSelected} />
        </>
    )
}
