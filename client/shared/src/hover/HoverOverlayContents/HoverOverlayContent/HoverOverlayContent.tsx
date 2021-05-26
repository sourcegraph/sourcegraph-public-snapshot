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
        return <span className="hover-overlay__content">{content.value}</span>
    }

    const markdownOrError = tryMarkdownRender(content.value)

    if (markdownOrError instanceof Error) {
        return (
            <div className={classNames('hover-overlay__icon', errorAlertClassName)}>
                {upperFirst(markdownOrError.message)}
            </div>
        )
    }

    return (
        <>
            {index !== 0 && <hr />}
            {aggregatedBadges.map(({ text, linkURL, hoverMessage }) => (
                <LinkOrSpan
                    key={text}
                    to={linkURL}
                    target="_blank"
                    rel="noopener noreferrer"
                    data-tooltip={hoverMessage}
                    className={classNames('hover-overlay__badge', 'test-hover-badge', badgeClassName)}
                >
                    {text}
                </LinkOrSpan>
            ))}
            <span
                className="hover-overlay__content test-tooltip-content"
                dangerouslySetInnerHTML={{ __html: markdownOrError }}
            />
        </>
    )
}
