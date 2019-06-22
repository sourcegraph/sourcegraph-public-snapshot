import Octicon, { GitPullRequest } from '@primer/octicons-react'
import React from 'react'

// The type definitions for Octicon's props erroneously omit className, so patch them to include
// className.
//
// TODO(sqs): Remove this when https://github.com/primer/octicons/pull/271 is merged.
declare module '@primer/octicons-react' {
    export interface OcticonProps {
        className: string
    }
}

export const ChangesetIcon: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <Octicon icon={GitPullRequest} className={className} />
)
