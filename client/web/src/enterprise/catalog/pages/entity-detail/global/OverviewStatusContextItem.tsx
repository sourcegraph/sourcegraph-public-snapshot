import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { ComponentStatusFields, ComponentStatusState } from '../../../../../graphql-operations'

import styles from './OverviewStatusContextItem.module.scss'

interface Props {
    statusContext: Omit<ComponentStatusFields['status']['contexts'][0], 'id'>
    className?: string
}

export const OverviewStatusContextItem: React.FunctionComponent<Props> = ({ statusContext, className, children }) => {
    const color = STATE_TO_COLOR[statusContext.state]
    return (
        <div className={classNames('d-flex align-items-start', className)}>
            <h4
                className={classNames(
                    `badge bg-transparent mb-0 mr-2 border border-${color} text-${color}`,
                    styles.title
                )}
            >
                {statusContext.targetURL ? (
                    <Link to={statusContext.targetURL} className={`d-block text-${color}`}>
                        {statusContext.title}
                    </Link>
                ) : (
                    statusContext.title
                )}
            </h4>
            <div>
                {statusContext.description || (statusContext.targetURL && !children) ? (
                    <div className="d-flex align-items-center">
                        {statusContext.description && (
                            <Markdown dangerousInnerHTML={renderMarkdown(statusContext.description)} />
                        )}
                        {statusContext.targetURL && !children && (
                            <span className="small ml-2">
                                <Link to={statusContext.targetURL} className="text-muted">
                                    Details
                                </Link>
                            </span>
                        )}
                    </div>
                ) : null}
                {children}
            </div>
        </div>
    )
}

export const STATE_TO_COLOR: Record<
    ComponentStatusState,
    'primary' | 'success' | 'warning' | 'danger' | 'info' | 'secondary'
> = {
    PENDING: 'secondary',
    EXPECTED: 'secondary',
    ERROR: 'danger',
    FAILURE: 'danger',
    INFO: 'primary',
    SUCCESS: 'success',
}
