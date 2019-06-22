import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { WithLinkPreviews } from '../../../../../../shared/src/components/linkPreviews/WithLinkPreviews'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'
import { LINK_PREVIEW_CLASS } from '../../../../components/linkPreviews/styles'
import { setElementTooltip } from '../../../../components/tooltip/Tooltip'

interface Props extends ExtensionsControllerProps {
    task: sourcegraph.Diagnostic

    className?: string
    location: H.Location
    history: H.History
}

/**
 * The description for a single task.
 */
export const TaskDescription: React.FunctionComponent<Props> = ({ task, className, ...props }) => (
    // TODO!(sqs): sanitize markdown
    <WithLinkPreviews
        dangerousInnerHTML={renderMarkdown(task.message)}
        extensionsController={props.extensionsController}
        setElementTooltip={setElementTooltip}
        linkPreviewContentClass={LINK_PREVIEW_CLASS}
    >
        {props => <Markdown {...props} />}
    </WithLinkPreviews>
)
