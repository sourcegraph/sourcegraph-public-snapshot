import classNames from 'classnames'
import React from 'react'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'

import { BatchChangeFields } from '../../graphql-operations'

interface DescriptionProps extends Pick<BatchChangeFields, 'description'> {
    className?: string
}

export const Description: React.FunctionComponent<DescriptionProps> = ({ description, className }) => (
    <div className={classNames('mb-3', className)}>
        <Markdown dangerousInnerHTML={renderMarkdown(description || '_No description_')} />
    </div>
)
