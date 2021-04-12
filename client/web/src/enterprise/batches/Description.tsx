import classNames from 'classnames'
import * as H from 'history'
import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { BatchChangeFields } from '../../graphql-operations'

interface DescriptionProps extends Pick<BatchChangeFields, 'description'> {
    history: H.History
    className?: string
}

export const Description: React.FunctionComponent<DescriptionProps> = ({ description, history, className }) => (
    <div className={classNames('mb-3', className)}>
        <Markdown dangerousInnerHTML={renderMarkdown(description || '_No description_')} history={history} />
    </div>
)
