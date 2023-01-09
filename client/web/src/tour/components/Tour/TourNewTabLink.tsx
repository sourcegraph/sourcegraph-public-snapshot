import { FunctionComponent } from 'react'

import { TourTaskStepType } from '@sourcegraph/shared/src/settings/temporary'
import { ButtonLink, Link } from '@sourcegraph/wildcard'

export interface NewTabLinkProps {
    step: TourTaskStepType
    variant: 'button' | 'link'
    className?: string
    to: string
    onClick: (
        event: React.MouseEvent<HTMLElement, MouseEvent> | React.KeyboardEvent<HTMLElement>,
        step: TourTaskStepType
    ) => void
}

export const TourNewTabLink: FunctionComponent<NewTabLinkProps> = ({ step, onClick, variant, to }) => {
    const commonLinkProps = {
        className: 'flex-grow-1',
        target: '_blank',
        rel: 'noopener noreferrer',
        to,
    }

    if (variant === 'button') {
        return (
            <ButtonLink variant="primary" {...commonLinkProps} onSelect={event => onClick(event, step)}>
                {step.label}
            </ButtonLink>
        )
    }

    return (
        <Link {...commonLinkProps} onClick={event => onClick(event, step)}>
            {step.label}
        </Link>
    )
}
