import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { isErrorLike } from '../../util/errors'
import { HoverOverlayBaseProps } from '../HoverOverlay.types'

import { HoverOverlayContent } from './HoverOverlayContent'

interface HoverOverlayContentsProps extends Pick<HoverOverlayBaseProps, 'hoverOrError'> {
    iconClassName?: string
    badgeClassName?: string
    errorAlertClassName?: string
}

export const HoverOverlayContents: React.FunctionComponent<HoverOverlayContentsProps> = props => {
    const { hoverOrError, iconClassName, errorAlertClassName, badgeClassName } = props

    if (hoverOrError === 'loading') {
        return (
            <div className="hover-overlay__loader-row">
                <LoadingSpinner className={iconClassName} />
            </div>
        )
    }

    if (isErrorLike(hoverOrError)) {
        return (
            <div className={classNames('hover-overlay__hover-error', errorAlertClassName)}>
                {upperFirst(hoverOrError.message)}
            </div>
        )
    }

    if (hoverOrError === undefined) {
        return null
    }

    if (hoverOrError === null || (hoverOrError.contents.length === 0 && hoverOrError.alerts?.length)) {
        return (
            // Show some content to give the close button space
            // and communicate to the user we couldn't find a hover.
            <em>No hover information available.</em>
        )
    }

    return (
        <>
            {hoverOrError.contents.map((content, index) => (
                <HoverOverlayContent
                    key={index}
                    index={index}
                    content={content}
                    aggregatedBadges={hoverOrError.aggregatedBadges}
                    errorAlertClassName={errorAlertClassName}
                    badgeClassName={badgeClassName}
                />
            ))}
        </>
    )
}
