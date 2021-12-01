import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { CatalogEntityStatusFields, CatalogEntityStatusState } from '../../../../../graphql-operations'

interface Props {
    statusContext: Omit<CatalogEntityStatusFields['status']['contexts'][0], 'id'>
    className?: string
}

export const OverviewStatusContextItem: React.FunctionComponent<Props> = ({ statusContext, className, children }) => {
    console.log('XX')
    const color = STATE_TO_COLOR[statusContext.state]
    return (
        <div className={classNames('d-flex align-items-start', className)}>
            <h4
                className={classNames(`badge bg-transparent mb-0 mr-2 border border-${color} text-${color}`)}
                // eslint-disable-next-line react/forbid-dom-props
                style={{ marginTop: '-1px' }}
            >
                {statusContext.targetURL ? (
                    <Link to={statusContext.targetURL} className={`text-${color}`}>
                        {statusContext.title}
                    </Link>
                ) : (
                    statusContext.title
                )}
            </h4>
            <div>
                {statusContext.description && (
                    <Markdown dangerousInnerHTML={renderMarkdown(statusContext.description)} />
                )}
                {children}
            </div>
        </div>
    )
}

const STATE_TO_COLOR: Record<
    CatalogEntityStatusState,
    'primary' | 'success' | 'warning' | 'danger' | 'info' | 'secondary'
> = {
    PENDING: 'secondary',
    EXPECTED: 'secondary',
    ERROR: 'danger',
    FAILURE: 'danger',
    INFO: 'primary',
    SUCCESS: 'success',
}
