import { mdiLoading, mdiPlus } from '@mdi/js'

import { useMutation } from '@sourcegraph/http-client'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { AssignOwnerResult, AssignOwnerVariables } from '../../../graphql-operations'

import { ASSIGN_OWNER } from './grapqlQueries'

export interface MakeOwnerButtonProps {
    onSuccess: () => Promise<any>
    onError: (e: Error) => void
    repoId: string
    path: string
    userId?: string
}

export const MakeOwnerButton: React.FC<MakeOwnerButtonProps> = ({ onSuccess, onError, repoId, path, userId }) => {
    const tooltipContent =
        userId === undefined
            ? 'Only ownership entries that are recognized as Sourcegraph users can be assigned ownership.'
            : null

    const [requestAssignOwner, { loading }] = useMutation<AssignOwnerResult, AssignOwnerVariables>(ASSIGN_OWNER, {})

    const assignOwner = async () => {
        if (userId !== undefined) {
            const result = await requestAssignOwner({
                variables: {
                    input: {
                        absolutePath: path,
                        assignedOwnerID: userId,
                        repoID: repoId,
                    },
                },
            })
            if (result.errors) {
                onError(new Error('Failed to make owner.'))
            } else {
                await onSuccess()
            }
        }
    }

    return (
        <Tooltip content={tooltipContent}>
            <Button onClick={assignOwner} variant="primary" outline={true} size="sm" disabled={userId === undefined}>
                <Icon aria-hidden={true} svgPath={loading ? mdiLoading : mdiPlus} />
                Make owner
            </Button>
        </Tooltip>
    )
}
