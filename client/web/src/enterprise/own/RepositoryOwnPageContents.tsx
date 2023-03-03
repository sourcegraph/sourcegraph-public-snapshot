import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiTrashCan, mdiUpload } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { H3, Text, Button, Icon, Code, Card, LoadingSpinner, ErrorAlert, Input } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import {
    AddIngestedCodeownersResult,
    AddIngestedCodeownersVariables,
    DeleteIngestedCodeownersResult,
    DeleteIngestedCodeownersVariables,
    GetIngestedCodeownersResult,
    GetIngestedCodeownersVariables,
    IngestedCodeowners,
    RepositoryFields,
    UpdateIngestedCodeownersResult,
    UpdateIngestedCodeownersVariables,
} from '../../graphql-operations'

import {
    ADD_INGESTED_CODEOWNERS_MUTATION,
    DELETE_INGESTED_CODEOWNERS_MUTATION,
    GET_INGESTED_CODEOWNERS_QUERY,
    UPDATE_INGESTED_CODEOWNERS_MUTATION,
} from './graphqlQueries'
import { IngestedFileViewer } from './IngestedFileViewer'
import { RepositoryOwnAreaPageProps } from './RepositoryOwnPage'

import styles from './RepositoryOwnPageContents.module.scss'

const MAX_FILE_SIZE_IN_BYTES = 10 * 1024 * 1024 // 10MB

export interface CodeownersIngestedFile {
    contents: string
    updatedAt: string
}

export const RepositoryOwnPageContents: React.FunctionComponent<
    Pick<RepositoryOwnAreaPageProps, 'repo' | 'authenticatedUser'>
> = ({ repo, authenticatedUser }) => {
    const isAdmin = authenticatedUser?.siteAdmin

    const { data, error, loading } = useQuery<GetIngestedCodeownersResult, GetIngestedCodeownersVariables>(
        GET_INGESTED_CODEOWNERS_QUERY,
        {
            variables: {
                repoID: repo.id,
            },
        }
    )

    const [codeownersIngestedFile, setCodeownersIngestedFile] = useState<IngestedCodeowners | null>(null)
    useEffect(() => {
        if (data?.node?.__typename === 'Repository') {
            if (data.node.ingestedCodeowners?.__typename === 'CodeownersIngestedFile') {
                setCodeownersIngestedFile(data.node.ingestedCodeowners)
            } else {
                setCodeownersIngestedFile(null)
            }
        }
    }, [data?.node])

    const [uploadError, setUploadError] = useState<ErrorLike | null>(null)
    const [addCodeonwersFile, addCodeonwersFileResult] = useMutation<
        AddIngestedCodeownersResult,
        AddIngestedCodeownersVariables
    >(ADD_INGESTED_CODEOWNERS_MUTATION)
    const [updateCodeownersFile, updateCodeownersFileResult] = useMutation<
        UpdateIngestedCodeownersResult,
        UpdateIngestedCodeownersVariables
    >(UPDATE_INGESTED_CODEOWNERS_MUTATION)

    const [deleteError, setDeleteError] = useState<ErrorLike | null>(null)
    const [deleteCodeownersFile] = useMutation<DeleteIngestedCodeownersResult, DeleteIngestedCodeownersVariables>(
        DELETE_INGESTED_CODEOWNERS_MUTATION
    )

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

            const upload = codeownersIngestedFile ? updateCodeownersFile : addCodeonwersFile
            const options = codeownersIngestedFile ? updateCodeownersFileResult : addCodeonwersFileResult

            upload({
                variables: {
                    repoID: repo.id,
                    contents,
                },
            })
                .then(result => {
                    if (result.data) {
                        if ('updateCodeownersFile' in result.data) {
                            setCodeownersIngestedFile(result.data.updateCodeownersFile)
                        } else {
                            setCodeownersIngestedFile(result.data.addCodeownersFile)
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
                    options.reset()
                })
        })
        reader.readAsText(file)
    }, [
        addCodeonwersFile,
        addCodeonwersFileResult,
        codeownersIngestedFile,
        repo.id,
        updateCodeownersFile,
        updateCodeownersFileResult,
    ])

    if (loading) {
        return (
            <div className="container d-flex justify-content-center mt-3">
                <LoadingSpinner /> Loading...
            </div>
        )
    }

    if (error) {
        return <ErrorAlert className="mt-3" error={error} prefix="Error loading ownership info for this repository" />
    }

    return (
        <>
            <Card className={styles.columns}>
                <div>
                    <H3>{isAdmin ? 'Upload a CODEOWNERS file' : 'Ask your site admin to upload a CODEOWNERS file'}</H3>
                    <Text>
                        {!isAdmin && 'A site admin can manually upload a CODEOWNERS file for this repository. '} Each
                        owner must be either a Sourcegraph username, a Sourcegraph team name, or an email address.
                    </Text>

                    {isAdmin && (
                        <>
                            <LoaderButton
                                icon={<Icon svgPath={mdiUpload} aria-hidden={true} className="mr-2" />}
                                label={codeownersIngestedFile ? 'Replace current file' : 'Upload file'}
                                loading={addCodeonwersFileResult.loading || updateCodeownersFileResult.loading}
                                variant="primary"
                                onClick={onUploadClicked}
                            />
                            {uploadError && (
                                <ErrorAlert className="mt-2" error={uploadError} prefix="Error uploading file:" />
                            )}
                            {/* Don't show the file input, the nicer-looking button will trigger it programmatically */}
                            <Input ref={fileInputRef} type="file" className="d-none" onChange={onFileSelected} />
                        </>
                    )}
                </div>

                <div className={styles.or}>
                    <div className={styles.orLine} />
                    <div className="py-2">or</div>
                    <div className={styles.orLine} />
                </div>

                <div>
                    <H3>Commit a CODEOWNERS file</H3>
                    <Text>
                        Add a <Code>CODEOWNERS</Code> file to the root of this repository. Owners must be{' '}
                        {getCodeHostName(repo)} usernames or email addresses.
                    </Text>
                    {codeownersIngestedFile && (
                        <Text className={styles.commitWarning}>
                            <em>
                                Any commited CODEOWNERS file in this repository will be ignored unless the uploaded file
                                is deleted.
                            </em>
                        </Text>
                    )}
                </div>
            </Card>

            {codeownersIngestedFile && (
                <div className="mt-5">
                    <H3>Uploaded CODEOWNERS file</H3>
                    <div className="d-flex align-items-baseline justify-content-between">
                        <Text>
                            The following CODEOWNERS file was uploaded to Sourcegraph{' '}
                            <Timestamp date={codeownersIngestedFile.updatedAt} />.
                        </Text>
                        <Button variant="danger" outline={true} className="ml-2">
                            <Icon svgPath={mdiTrashCan} aria-hidden={true} className="mr-2" />
                            Delete uploaded file
                        </Button>
                    </div>
                    <IngestedFileViewer contents={codeownersIngestedFile.contents} />
                </div>
            )}
        </>
    )
}

const getCodeHostName = (repo: RepositoryFields): string => {
    const externalServiceKind = repo.externalURLs[0]?.serviceKind

    switch (externalServiceKind) {
        case 'GITHUB':
            return 'GitHub'
        case 'GITLAB':
            return 'GitLab'
        default:
            return 'code host'
    }
}
