import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { HoverMerged } from '../../../api/client/types/hover'
import { LinkOrSpan } from '../../../components/LinkOrSpan'
import { asError } from '../../../util/errors'
import { renderMarkdown } from '../../../util/markdown'

interface HoverOverlayContentProps {
    content: HoverMerged['contents'][number]
    aggregatedBadges: HoverMerged['aggregatedBadges']
    index: number
    badgeClassName?: string
    errorAlertClassName?: string
}

function tryMarkdownRender(content: string): string | Error {
    try {
        return renderMarkdown(content)
    } catch (error) {
        return asError(error)
    }
}

export const HoverOverlayContent: React.FunctionComponent<HoverOverlayContentProps> = props => {
    const { content, aggregatedBadges = [], index, errorAlertClassName, badgeClassName } = props

    if (content.kind !== 'markdown') {
        return (
            <span className="hover-overlay__content">
                <p>{content.value}</p>
            </span>
        )
    }

    const markdownOrError = tryMarkdownRender(content.value)

    if (markdownOrError instanceof Error) {
        return (
            <div className={classNames('hover-overlay__hover-error', errorAlertClassName)}>
                {upperFirst(markdownOrError.message)}
            </div>
        )
    }

    return (
        <>
            {index !== 0 && <hr />}
            {aggregatedBadges.map(({ text, linkURL, hoverMessage }) => (
                <small key={text} className="hover-overlay__badge">
                    <LinkOrSpan
                        to={linkURL}
                        target="_blank"
                        rel="noopener noreferrer"
                        data-tooltip={hoverMessage}
                        className={classNames('test-hover-badge', badgeClassName, 'hover-overlay__badge-label')}
                    >
                        {text}
                    </LinkOrSpan>
                </small>
            ))}
            <span
                className="hover-overlay__content test-tooltip-content"
                dangerouslySetInnerHTML={{ __html: markdownOrError }}
            />
        </>
    )
}
