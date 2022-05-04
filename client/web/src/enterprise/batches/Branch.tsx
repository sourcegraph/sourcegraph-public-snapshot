import React from 'react'

import classNames from 'classnames'
import AccountQuestionIcon from 'mdi-react/AccountQuestionIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'

import { Badge, Icon, BadgeProps } from '@sourcegraph/wildcard'

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
        as={deleted ? 'del' : undefined}
    >
        {!forkTarget || forkTarget.namespace === null ? (
            name
        ) : (
            <>
                <Icon className="mr-1" as={SourceForkIcon} />
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
    <div className="d-block d-sm-inline-block">
        <Branch name={baseRef} />
        <span className="p-1">&larr;</span>
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
        return (
            <>
                <Icon
                    data-tooltip="This branch will be pushed to a user fork. If you have configured a credential for yourself in the Batch Changes settings, this will be a fork in your code host account; otherwise the fork will be in the code host account associated with the site credential used to open changesets."
                    as={AccountQuestionIcon}
                />
                :
            </>
        )
    }

    return <>{target.namespace}:</>
}
