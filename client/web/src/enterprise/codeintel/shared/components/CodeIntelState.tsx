import React from 'react'

import classNames from 'classnames'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

import { CodeIntelStateIcon } from './CodeIntelStateIcon'
import { CodeIntelStateLabel } from './CodeIntelStateLabel'

export interface CodeIntelStateProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'state' | 'placeInQueue'>
    className?: string
}

const iconClassNames = 'm-0 text-nowrap justify-content-center'

export const CodeIntelState: React.FunctionComponent<React.PropsWithChildren<CodeIntelStateProps>> = ({
    node,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <CodeIntelStateIcon state={node.state} />
        <CodeIntelStateLabel state={node.state} placeInQueue={node.placeInQueue} className="mt-2 ml-2" />
    </div>
)
