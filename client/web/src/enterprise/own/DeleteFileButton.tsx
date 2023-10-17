import { useCallback, useState } from 'react'

import { mdiTrashCan } from '@mdi/js'

import { type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, ErrorAlert, H3, Icon, Modal } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type {
    DeleteIngestedCodeownersResult,
    DeleteIngestedCodeownersVariables,
    RepositoryFields,
} from '../../graphql-operations'

import { DELETE_INGESTED_CODEOWNERS_MUTATION } from './graphqlQueries'

interface DeleteFileButtonProps {
    onComplete: () => void
    repo: RepositoryFields
}

export const DeleteFileButton: React.FunctionComponent<DeleteFileButtonProps> = ({ repo, onComplete }) => {
    const [showModal, setShowModal] = useState(false)

    const [deleteError, setDeleteError] = useState<ErrorLike | null>(null)
    const [deleteCodeownersFile, { loading }] = useMutation<
        DeleteIngestedCodeownersResult,
        DeleteIngestedCodeownersVariables
    >(DELETE_INGESTED_CODEOWNERS_MUTATION)

    const onDeleteClicked = useCallback(() => {
        deleteCodeownersFile({ variables: { repoID: repo.id } })
            .then(() => {
                onComplete()
                setShowModal(false)
            })
            .catch(error => {
                if (isErrorLike(error)) {
                    setDeleteError(error)
                } else {
                    setDeleteError(new Error('Unknown error'))
                }
            })
    }, [deleteCodeownersFile, onComplete, repo.id])

    return (
        <>
            <Button variant="danger" outline={true} className="ml-2" onClick={() => setShowModal(true)}>
                <Icon svgPath={mdiTrashCan} aria-hidden={true} className="mr-2" />
                Delete uploaded file
            </Button>

            {showModal && (
                <Modal onDismiss={() => setShowModal(false)} aria-labelledby="delete-codeowners">
                    <H3 id="delete-codeowners">Are you sure you want to delete this uploaded CODEOWNERS file?</H3>
                    <strong className="d-block text-danger my-3">Deleting is irreversible.</strong>

                    {deleteError && <ErrorAlert className="mt-2" error={deleteError} prefix="Error deleting file: " />}
                    <div className="d-flex justify-content-end pt-1">
                        <Button
                            disabled={loading}
                            className="mr-2"
                            onClick={() => setShowModal(false)}
                            outline={true}
                            variant="secondary"
                        >
                            Cancel
                        </Button>
                        <LoaderButton
                            variant="danger"
                            loading={loading}
                            onClick={onDeleteClicked}
                            label="Delete uploaded file"
                        />
                    </div>
                </Modal>
            )}
        </>
    )
}
