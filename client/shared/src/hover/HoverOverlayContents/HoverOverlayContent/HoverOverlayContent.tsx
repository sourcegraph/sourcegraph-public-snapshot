import React from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import type { HoverMerged } from '@sourcegraph/client-api'
import { asError, renderMarkdown } from '@sourcegraph/common'
import { Alert, type AlertProps, Badge, Text } from '@sourcegraph/wildcard'

import hoverOverlayStyle from '../../HoverOverlay.module.scss'
import hoverOverlayContentsStyle from '../../HoverOverlayContents.module.scss'
import style from './HoverOverlayContent.module.scss'

interface HoverOverlayContentProps {
    content: HoverMerged['contents'][number]
    aggregatedBadges: HoverMerged['aggregatedBadges']
    index: number
    /**
     * Allows custom styles
     * Primarily used to inherit different styles for use on a code host.
     */
    badgeClassName?: string
    errorAlertClassName?: string
    errorAlertVariant?: AlertProps['variant']
    contentClassName?: string
}

function tryMarkdownRender(content: string): string | Error {
    try {
        return renderMarkdown(content)
    } catch (error) {
        return asError(error)
    }
}

export const HoverOverlayContent: React.FunctionComponent<
    React.PropsWithChildren<HoverOverlayContentProps>
> = props => {
    const { content, aggregatedBadges = [], index, errorAlertClassName, errorAlertVariant, badgeClassName } = props

    if (content.kind !== 'markdown') {
        return (
            <span
                data-testid="hover-overlay-content"
                className={classNames(style.hoverOverlayContent, hoverOverlayContentsStyle.hoverOverlayContent)}
            >
                <Text>{content.value}</Text>
            </span>
        )
    }

    const markdownOrError = tryMarkdownRender(content.value)

    if (markdownOrError instanceof Error) {
        return (
            <Alert
                className={classNames(hoverOverlayStyle.hoverError, errorAlertClassName)}
                variant={errorAlertVariant}
            >
                {upperFirst(markdownOrError.message)}
            </Alert>
        )
    }

    return (
        <>
            {index !== 0 && <hr />}
            {aggregatedBadges.map(({ text, linkURL, hoverMessage }) => (
                <small key={text} className={classNames(hoverOverlayStyle.badge)}>
                    <Badge
                        variant="secondary"
                        small={true}
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
