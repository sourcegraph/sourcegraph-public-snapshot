import { useCallback, useState } from 'react'

import { mdiTrashCan } from '@mdi/js'

import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, ErrorAlert, Icon } from '@sourcegraph/wildcard'

import {
    DeleteIngestedCodeownersResult,
    DeleteIngestedCodeownersVariables,
    RepositoryFields,
} from '../../graphql-operations'

import { DELETE_INGESTED_CODEOWNERS_MUTATION } from './graphqlQueries'
import { LoaderButton } from '../../components/LoaderButton'

interface DeleteFileButtonProps {
    onComplete: () => void
    repo: RepositoryFields
}

export const DeleteFileButton: React.FunctionComponent<DeleteFileButtonProps> = ({ repo, onComplete }) => {
    const [deleteError, setDeleteError] = useState<ErrorLike | null>(null)
    const [deleteCodeownersFile, { loading, reset }] = useMutation<
        DeleteIngestedCodeownersResult,
        DeleteIngestedCodeownersVariables
    >(DELETE_INGESTED_CODEOWNERS_MUTATION)

    const onDeleteClicked = useCallback(() => {
        deleteCodeownersFile({variables: {repoID: repo.id}})
        .then(() => {
            onComplete()
        })
        .catch(error => {
            if (isErrorLike(error)) {
                setDeleteError(error)
            } else {
                setDeleteError(new Error('Unknown error'))
            }
        })
        .finally(() => {
            reset()
        }

    }, [])

    return (
        <>
            <LoaderButton variant="danger" outline={true} className="ml-2" loading={loading} onClick={onDeleteClicked}>
                <Icon svgPath={mdiTrashCan} aria-hidden={true} className="mr-2" />
                Delete uploaded file
            </LoaderButton>
            {deleteError && <ErrorAlert className="mt-2" error={deleteError} prefix="Error deleting file: " />}
        </>
    )
}
