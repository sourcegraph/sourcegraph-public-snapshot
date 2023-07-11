import { FC, useEffect, useRef, useCallback, useState } from 'react'

import { mdiCloudUpload } from '@mdi/js'

import { ErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Label,
    Button,
    Container,
    Link,
    PageHeader,
    Text,
    Icon,
    Input,
    ErrorAlert,
    LoadingSpinner,
    Alert,
} from '@sourcegraph/wildcard'

import { CodyColorIcon } from '../../../../cody/chat/CodyPageIcon'
import { PageTitle } from '../../../../components/PageTitle'

import styles from './EmbeddingUploadPage.module.scss'

const MAX_FILE_SIZE_IN_BYTES = 10 * 1024 * 1024 // 10MB

export interface EmbeddingUploadPageProps extends TelemetryProps {}

export const EmbeddingUploadPage: FC<EmbeddingUploadPageProps> = ({ telemetryService }) => {
    const [uploadError, setUploadError] = useState<ErrorLike | null>(null)
    const [fileToUpload, setFileToUpload] = useState<File | null>(null)
    const [isLoading, setIsLoading] = useState<boolean>(false)
    const [isSuccess, setIsSuccess] = useState<boolean>(false)

    useEffect(() => {
        telemetryService.logPageView('EmbeddingUploadPage')
    }, [telemetryService])

    const fileInputRef = useRef<HTMLInputElement | null>(null)

    const upload = useCallback(async () => {
        try {
            setIsLoading(true)
            setIsSuccess(false)
            setUploadError(null)

            if (fileToUpload) {
                const formdata = new FormData()
                formdata.append('archive', fileToUpload, fileToUpload.name)
                const resp = await fetch('/.api/files/embeddings', {
                    method: 'POST',
                    body: formdata,
                    headers: {
                        ...window.context.xhrHeaders,
                    },
                })

                if (!resp.ok) {
                    const err = await resp.text()
                    setUploadError(new Error(`Error uploading file: ${err}`))
                    return
                }

                setIsSuccess(true)
                setUploadError(null)
                setFileToUpload(null)
                return
            }
            setUploadError(new Error('No file to upload.'))
        } catch (error) {
            setUploadError(error)
        } finally {
            setIsLoading(false)
        }
    }, [fileToUpload])

    const onFileSelected = useCallback(() => {
        setUploadError(null)
        const file = fileInputRef.current?.files?.[0]
        if (!file) {
            return
        }

        if (file.size > MAX_FILE_SIZE_IN_BYTES) {
            setUploadError(new Error(`File must not exceed ${MAX_FILE_SIZE_IN_BYTES / 1024 / 1024}MB in size.`))
            return
        }

        setFileToUpload(file)
    }, [])

    return (
        <>
            <PageTitle title="Embedding Third-party Data Upload" />
            <PageHeader
                headingElement="h2"
                path={[
                    { icon: CodyColorIcon, text: 'Cody' },
                    {
                        text: 'Upload',
                    },
                ]}
                description={
                    <>
                        Rules that control keeping embeddings up-to-date. See the{' '}
                        <Link target="_blank" to="/help/cody/explanations/policies">
                            documentation
                        </Link>{' '}
                        for more details.
                    </>
                }
                className="mb-3"
            />

            <Container>
                {uploadError && <ErrorAlert className="mt-2" error={uploadError} prefix="Error uploading file:" />}
                {isSuccess && (
                    <Alert variant="success">
                        <Text>Data successfully uploaded.</Text>
                    </Alert>
                )}
                <Label htmlFor="file-upload" className={styles.uploadContainer}>
                    <Input
                        id="file-upload"
                        ref={fileInputRef}
                        type="file"
                        className="d-none"
                        onChange={onFileSelected}
                        accept=".zip"
                        disabled={isLoading}
                    />
                    {isLoading ? (
                        <>
                            <LoadingSpinner />
                            <Text>Uploading ...</Text>
                        </>
                    ) : (
                        <>
                            <Icon
                                className={styles.uploadIcon}
                                svgPath={mdiCloudUpload}
                                inline={false}
                                aria-label="Upload archive"
                            />
                            <Text>{fileToUpload ? fileToUpload.name : 'Click to upload an archive to embed.'}</Text>
                        </>
                    )}
                </Label>
                <Button variant="primary" disabled={fileToUpload === null || isLoading} onClick={upload}>
                    Upload
                </Button>
            </Container>
        </>
    )
}
