import { mdiPlus } from '@mdi/js'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

export interface MakeOwnerButtonProps {
    onSuccess: () => Promise<any>
    onError: (e: Error) => void
    repoId: string
    path: string
    userId?: string
}

export const MakeOwnerButton: React.FC<MakeOwnerButtonProps> = ({ userId }) => {
    const tooltipContent =
        userId === undefined
            ? 'Only ownership entries that are recognized as Sourcegraph users can be assigned ownership.'
            : null

    return (
        <Tooltip content={tooltipContent}>
            <Button variant="primary" outline={true} size="sm" disabled={userId === undefined}>
                <Icon aria-hidden={true} svgPath={mdiPlus} />
                Make owner
            </Button>
        </Tooltip>
    )
}
