import React from 'react'

import { mdiLoading, mdiMinusThick } from '@mdi/js'

import { type ErrorLike, asError } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import type {
    RemoveAssignedOwnerResult,
    RemoveAssignedOwnerVariables,
    RemoveAssignedTeamResult,
    RemoveAssignedTeamVariables,
} from '../../../graphql-operations'

import { REMOVE_ASSIGNED_OWNER, REMOVE_ASSIGNED_TEAM } from './grapqlQueries'

import styles from './RemoveOwnerButton.module.scss'

export interface RemoveOwnerButtonProps {
    onSuccess: () => Promise<any>
    onError: (e: Error) => void
    repoId: string
    path: string
    userID?: string
    teamID?: string
    isDirectAssigned: boolean
}

export const RemoveOwnerButton: React.FC<RemoveOwnerButtonProps> = ({
    onSuccess,
    onError,
    repoId,
    path,
    userID,
    teamID,
    isDirectAssigned,
}) => {
    const tooltipContent = !isDirectAssigned
        ? 'Ownership can only be modified at the same direct path as it was assigned.'
        : 'Remove ownership'

    const [removeAssignedOwner, { loading: removeLoading }] = useMutation<
        RemoveAssignedOwnerResult,
        RemoveAssignedOwnerVariables
    >(REMOVE_ASSIGNED_OWNER, {})
    const [removeAssignedTeam, { loading: removeTeamLoading }] = useMutation<
        RemoveAssignedTeamResult,
        RemoveAssignedTeamVariables
    >(REMOVE_ASSIGNED_TEAM, {})

    const createInputObject = (id: string): any => ({
        variables: {
            input: {
                absolutePath: path,
                assignedOwnerID: id,
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

    const removeOwner: () => Promise<void> = async () => {
        if (userID) {
            await removeAssignedOwner(createInputObject(userID))
        } else if (teamID) {
            await removeAssignedTeam(createInputObject(teamID))
        }
    }

    return (
        <Tooltip content={tooltipContent}>
            <Button
                variant="icon"
                aria-label="Remove this ownership"
                onClick={removeOwner}
                outline={false}
                size="sm"
                disabled={!isDirectAssigned}
            >
                <Icon
                    className={styles.minusIcon}
                    aria-hidden={true}
                    svgPath={removeLoading || removeTeamLoading ? mdiLoading : mdiMinusThick}
                />
            </Button>
        </Tooltip>
    )
}
