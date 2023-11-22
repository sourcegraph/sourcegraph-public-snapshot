import type { FunctionComponent } from 'react'

import type { TourTaskStepType } from '@sourcegraph/shared/src/settings/temporary'
import { ButtonLink, Link } from '@sourcegraph/wildcard'

export interface NewTabLinkProps {
    step: TourTaskStepType
    variant: 'button' | 'link'
    className?: string
    to: string
    onClick: (step: TourTaskStepType) => void
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
            <ButtonLink variant="primary" {...commonLinkProps} onSelect={() => onClick(step)}>
                {step.label}
            </ButtonLink>
        )
    }

    return (
        <Link {...commonLinkProps} onClick={() => onClick(step)}>
            {step.label}
        </Link>
    )
}
