import { mdiPlus } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

export interface MakeOwnerButtonProps {
    onSuccess: () => Promise
}

export const MakeOwnerButton: React.FC<MakeOwnerButtonProps> = ({}) => {
    return (
        <Button variant={'primary'} outline={true} size={'sm'}>
            <Icon svgPath={mdiPlus} />
            Make owner
        </Button>
    )
}
