import classNames from 'classnames'
import AccountQuestionIcon from 'mdi-react/AccountQuestionIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import React from 'react'

import { Badge } from '@sourcegraph/wildcard'
import { BadgeProps } from '@sourcegraph/wildcard/src/components/Badge'

export interface ForkTarget {
    pushUser: boolean
    namespace?: string | null
}

export interface BranchProps extends Pick<BadgeProps, 'variant'> {
    className?: string
    deleted?: boolean
    forkTarget?: ForkTarget | null
    name: string
}

export const Branch: React.FunctionComponent<BranchProps> = ({ className, deleted, forkTarget, name, variant }) => (
    <Badge
        variant={variant !== undefined ? variant : deleted ? 'danger' : 'secondary'}
        className={classNames('text-monospace', className)}
        as={deleted ? 'del' : undefined}
    >
        {!forkTarget ? (
            name
        ) : (
            <>
                <SourceForkIcon className="icon-inline mr-1" />
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

export const BranchMerge: React.FunctionComponent<BranchMergeProps> = ({ baseRef, forkTarget, headRef }) => (
    <div className="d-block d-sm-inline-block">
        <Branch name={baseRef} />
        <span className="p-1">&larr;</span>
        <Branch name={headRef} forkTarget={forkTarget} />
    </div>
)

interface BranchNamespaceProps {
    target: ForkTarget
}

const BranchNamespace: React.FunctionComponent<BranchNamespaceProps> = ({ target }) => {
    if (!target) {
        return <></>
    }

    if (target.pushUser) {
        return (
            <>
                <AccountQuestionIcon
                    className="icon-inline"
                    data-tooltip="This branch will be pushed to a user fork. If you have configured a credential for yourself in the Batch Changes settings, this will be a fork in your code host account; otherwise the fork will be in the code host account associated with the site credential used to open changesets."
                />
                :
            </>
        )
    }

    return <>{target.namespace}:</>
}
