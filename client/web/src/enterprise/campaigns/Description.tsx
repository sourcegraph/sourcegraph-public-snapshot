import * as H from 'history'
import React from 'react'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { CampaignFields } from '../../graphql-operations'
import classNames from 'classnames'

interface DescriptionProps extends Pick<CampaignFields, 'description'> {
    history: H.History
    className?: string
}

export const Description: React.FunctionComponent<DescriptionProps> = ({ description, history, className }) => (
    <div className={classNames('mb-3', className)}>
        <Markdown dangerousInnerHTML={renderMarkdown(description || '_No description_')} history={history} />
    </div>
)
