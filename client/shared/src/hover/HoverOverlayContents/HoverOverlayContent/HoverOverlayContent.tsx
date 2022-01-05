import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { asError } from '@sourcegraph/common'
import { Badge } from '@sourcegraph/wildcard'

import { HoverMerged } from '../../../api/client/types/hover'
import { renderMarkdown } from '../../../util/markdown'
import hoverOverlayStyle from '../../HoverOverlay.module.scss'
import hoverOverlayContentsStyle from '../../HoverOverlayContents.module.scss'

import style from './HoverOverlayContent.module.scss'

interface HoverOverlayContentProps {
    content: HoverMerged['contents'][number]
    aggregatedBadges: HoverMerged['aggregatedBadges']
    index: number
    /**
     * Use our branded badge. For use in Sourcegraph branded applications.
     */
    useBrandedBadge?: boolean
    /**
     * Allows custom styles
     * Primarily used to inherit different styles for use on a code host.
     */
    badgeClassName?: string
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
    const { content, aggregatedBadges = [], index, errorAlertClassName, badgeClassName, useBrandedBadge } = props

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
                        branded={Boolean(useBrandedBadge)}
                        {...(useBrandedBadge && {
                            variant: 'secondary',
                            small: true,
                        })}
                        className={classNames('test-hover-badge', badgeClassName, hoverOverlayStyle.badgeLabel)}
                        href={linkURL}
                        tooltip={hoverMessage}
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
