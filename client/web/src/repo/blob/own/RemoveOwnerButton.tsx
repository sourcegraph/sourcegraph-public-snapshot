import React from 'react'

import { mdiDelete, mdiLoading } from '@mdi/js'

import { ErrorLike, asError } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { RemoveAssignedOwnerResult, RemoveAssignedOwnerVariables } from '../../../graphql-operations'

import { REMOVE_ASSIGNED_OWNER } from './grapqlQueries'

export interface RemoveOwnerButtonProps {
    onSuccess: () => Promise<any>
    onError: (e: Error) => void
    repoId: string
    path: string
    userId?: string
    isDirectAssigned: boolean
}

export const RemoveOwnerButton: React.FC<RemoveOwnerButtonProps> = ({
    onSuccess,
    onError,
    repoId,
    path,
    userId,
    isDirectAssigned,
}) => {
    const tooltipContent = !isDirectAssigned
        ? 'Ownership can only be modified at the same direct path as it was assigned.'
        : 'Remove ownership'

    const [removeAssignedOwner, { loading }] = useMutation<RemoveAssignedOwnerResult, RemoveAssignedOwnerVariables>(
        REMOVE_ASSIGNED_OWNER,
        {}
    )

    const removeOwner: () => Promise<void> = async () => {
        if (userId) {
            await removeAssignedOwner({
                variables: {
                    input: {
                        absolutePath: path,
                        assignedOwnerID: userId,
                        repoID: repoId,
                    },
                },
                onCompleted: async () => {
                    await onSuccess()
                },
                onError: (errors: ErrorLike) => {
                    onError(asError(errors))
                },
            })
        }
    }

    return (
        <Tooltip content={tooltipContent}>
            <Button
                variant="danger"
                className="ml-2"
                aria-label="Remove this ownership"
                onClick={removeOwner}
                outline={false}
                size="sm"
                disabled={!isDirectAssigned}
            >
                <Icon color="white" aria-hidden={true} svgPath={loading ? mdiLoading : mdiDelete} />
                Remove owner
            </Button>
        </Tooltip>
    )
}
