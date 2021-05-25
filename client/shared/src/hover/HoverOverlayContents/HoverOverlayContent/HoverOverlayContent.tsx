import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { HoverMerged } from '../../../api/client/types/hover'
import { LinkOrSpan } from '../../../components/LinkOrSpan'
import { asError } from '../../../util/errors'
import { renderMarkdown } from '../../../util/markdown'
import { useRedesignToggle } from '../../../util/useRedesignToggle'

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

    const [isRedesignEnabled] = useRedesignToggle()

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

    const BadgeWrapper = isRedesignEnabled ? 'small' : React.Fragment

    return (
        <>
            {index !== 0 && <hr />}
            {aggregatedBadges.map(({ text, linkURL, hoverMessage }) => (
                // In the redesign version wrapper has `badge` styles.
                // In the pre-redesign version wrapper is a React.Fragment, so we don't need a `className`.
                <BadgeWrapper {...(isRedesignEnabled && { className: 'hover-overlay__badge' })} key={text}>
                    <LinkOrSpan
                        to={linkURL}
                        target="_blank"
                        rel="noopener noreferrer"
                        data-tooltip={hoverMessage}
                        className={classNames(
                            'test-hover-badge',
                            badgeClassName,
                            // In the pre-redesign version `LinkOrSpan` has `badge` styles.
                            isRedesignEnabled ? 'hover-overlay__badge-label' : 'hover-overlay__badge'
                        )}
                    >
                        {text}
                    </LinkOrSpan>
                </BadgeWrapper>
            ))}
            <span
                className="hover-overlay__content test-tooltip-content"
                dangerouslySetInnerHTML={{ __html: markdownOrError }}
            />
        </>
    )
}
