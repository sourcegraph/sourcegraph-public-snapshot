import * as React from 'react'

import { Button } from '@sourcegraph/wildcard'

export const ShowMoreButton: React.FunctionComponent<
    React.PropsWithChildren<{
        onClick: () => void
        className?: string
        dataTestid?: string
    }>
> = ({ onClick, className, dataTestid }) => (
    <div className="text-center py-3">
        <Button className={className} onClick={onClick} data-testid={dataTestid} variant="link">
            Show more
        </Button>
    </div>
)
