import type { FC } from 'react'

import { Button, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

export type SubmitButtonState = 'disabled' | 'loading'

export interface SubmitButtonProps {
    onClick: () => void
    state?: SubmitButtonState
}

export const SubmitButton: FC<React.PropsWithChildren<SubmitButtonProps>> = ({ children, onClick, state }) => {
    if (state === 'loading') {
        return (
            <Button disabled={true} type="submit" variant="primary">
                <LoadingSpinner />
            </Button>
        )
    }

    if (state === 'disabled') {
        return (
            <Tooltip content="At least one event type must be selected.">
                <Button type="submit" variant="primary" disabled={true}>
                    {children}
                </Button>
            </Tooltip>
        )
    }

    return (
        <Button
            type="submit"
            variant="primary"
            onClick={event => {
                event.preventDefault()
                onClick()
            }}
        >
            {children}
        </Button>
    )
}
