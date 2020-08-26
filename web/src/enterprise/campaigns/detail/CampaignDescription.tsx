import * as H from 'history'
import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { CampaignFields } from '../../../graphql-operations'
import classNames from 'classnames'

interface CampaignDescriptionProps extends Pick<CampaignFields, 'description'> {
    history: H.History
    className?: string
}

export const CampaignDescription: React.FunctionComponent<CampaignDescriptionProps> = ({
    description,
    history,
    className,
}) => (
    <div className={classNames(className)}>
        <h3>Campaign description</h3>
        <hr />
        <div className="pl-3 py-3">
            <Markdown dangerousInnerHTML={renderMarkdown(description || '_No description_')} history={history} />
        </div>
    </div>
)
