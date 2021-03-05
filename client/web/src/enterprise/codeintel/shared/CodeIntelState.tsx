import classNames from 'classnames'
import React from 'react'
import { LsifIndexFields, LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelStateIcon } from './CodeIntelStateIcon'
import { CodeIntelStateLabel } from './CodeIntelStateLabel'

export interface CodeIntelStateProps {
    node: LsifUploadFields | LsifIndexFields
    className?: string
}

const iconClassNames = 'm-0 text-nowrap d-flex flex-column align-items-center justify-content-center'

export const CodeIntelState: React.FunctionComponent<CodeIntelStateProps> = ({ node, className }) => (
    <div className={classNames(iconClassNames, className)}>
        <CodeIntelStateIcon state={node.state} />
        <CodeIntelStateLabel state={node.state} placeInQueue={node.placeInQueue} className="mt-2" />
    </div>
)
