import React from 'react'

import { mdiSourceFork, mdiAccountQuestion } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Badge, Icon, type BadgeProps, Tooltip } from '@sourcegraph/wildcard'

export interface ForkTarget {
    pushUser: boolean
    namespace: string | null
}

export interface BranchProps extends Pick<BadgeProps, 'variant'> {
    className?: string
    deleted?: boolean
    forkTarget?: ForkTarget | null
    name: string
}

export const Branch: React.FunctionComponent<React.PropsWithChildren<BranchProps>> = ({
    className,
    deleted,
    forkTarget,
    name,
    variant,
}) => (
    <Badge
        variant={variant !== undefined ? variant : deleted ? 'danger' : 'secondary'}
        className={classNames('text-monospace', className)}
        as={deleted ? 'del' : 'span'}
        aria-label={deleted ? 'Deleted' : ''}
        pill={true}
    >
        {!forkTarget || forkTarget.namespace === null ? (
            name
        ) : (
            <>
                <Icon aria-label="fork" className="mr-1" svgPath={mdiSourceFork} />
                <BranchNamespace target={forkTarget} />
                {name}
            </>
        )}
    </Badge>
)

export interface BranchMergeProps {
    baseRef: string
    forkTarget?: ForkTarget | null
    headRef: string
}

export const BranchMerge: React.FunctionComponent<React.PropsWithChildren<BranchMergeProps>> = ({
    baseRef,
    forkTarget,
    headRef,
}) => (
    // Relative positioning needed to avoid VisuallyHidden creating a double layer scrollbar in Chrome.
    // Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
    <div className="d-block d-sm-inline-block position-relative">
        <VisuallyHidden>Request to merge commit into</VisuallyHidden>
        <Branch name={baseRef} />
        <Icon as="span" inline={false} className="p-1" aria-label="from">
            &larr;
        </Icon>
        <Branch name={headRef} forkTarget={forkTarget} />
    </div>
)

interface BranchNamespaceProps {
    target: ForkTarget
}

const BranchNamespace: React.FunctionComponent<React.PropsWithChildren<BranchNamespaceProps>> = ({ target }) => {
    if (!target) {
        return <></>
    }

    if (target.pushUser) {
        const iconLabel =
            'This branch will be pushed to a user fork. If you have configured a credential for yourself in the Batch Changes settings, this will be a fork in your code host account; otherwise the fork will be in the code host account associated with the site credential used to open changesets.'
        return (
            <>
                <Tooltip content={iconLabel}>
                    <Icon aria-label={iconLabel} svgPath={mdiAccountQuestion} />
                </Tooltip>
                :
            </>
        )
    }

    return <>{target.namespace}:</>
}
