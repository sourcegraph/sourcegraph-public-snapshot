import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

import { HoverMerged } from '../../../api/client/types/hover'
import { asError } from '../../../util/errors'
import { renderMarkdown } from '../../../util/markdown'
import hoverOverlayStyle from '../../HoverOverlay.module.scss'
import hoverOverlayContentsStyle from '../../HoverOverlayContents.module.scss'

import style from './HoverOverlayContent.module.scss'

interface HoverOverlayContentProps {
    content: HoverMerged['contents'][number]
    aggregatedBadges: HoverMerged['aggregatedBadges']
    index: number
    /**
     * Allows usage on other code hosts.
     * We can inherit the badge styles of the code host rather than use our branded styles.
     */
    customBadgeClassName?: string
    errorAlertClassName?: string
    contentClassName?: string
}

function tryMarkdownRender(content: string): string | Error {
    try {
        return renderMarkdown(content)
    } catch (error) {
        return asError(error)
    }
}

export const HoverOverlayContent: React.FunctionComponent<HoverOverlayContentProps> = props => {
    const { content, aggregatedBadges = [], index, errorAlertClassName, customBadgeClassName } = props

    if (content.kind !== 'markdown') {
        return (
            <span
                data-testid="hover-overlay-content"
                className={classNames(style.hoverOverlayContent, hoverOverlayContentsStyle.hoverOverlayContent)}
            >
                <p>{content.value}</p>
            </span>
        )
    }

    const markdownOrError = tryMarkdownRender(content.value)

    if (markdownOrError instanceof Error) {
        return (
            <div className={classNames(hoverOverlayStyle.hoverError, errorAlertClassName)}>
                {upperFirst(markdownOrError.message)}
            </div>
        )
    }

    return (
        <>
            {index !== 0 && <hr />}
            {aggregatedBadges.map(({ text, linkURL, hoverMessage }) => (
                <small key={text} className={classNames(hoverOverlayStyle.badge)}>
                    <Badge
                        unstyled={Boolean(customBadgeClassName)}
                        className={classNames('test-hover-badge', customBadgeClassName, hoverOverlayStyle.badgeLabel)}
                        href={linkURL}
                        tooltip={hoverMessage}
                        variant="secondary"
                        small={true}
                    >
                        {text}
                    </Badge>
                </small>
            ))}
            <span
                data-testid="hover-overlay-content"
                className={classNames(
                    style.hoverOverlayContent,
                    hoverOverlayContentsStyle.hoverOverlayContent,
                    props.contentClassName,
                    'test-tooltip-content'
                )}
                dangerouslySetInnerHTML={{ __html: markdownOrError }}
            />
        </>
    )
}
