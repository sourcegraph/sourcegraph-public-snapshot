import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { isErrorLike } from '../../util/errors'
import { useRedesignToggle } from '../../util/useRedesignToggle'
import { HoverOverlayBaseProps } from '../HoverOverlay.types'

import { HoverOverlayContent } from './HoverOverlayContent'

interface HoverOverlayContentsProps extends Pick<HoverOverlayBaseProps, 'hoverOrError'> {
    iconClassName?: string
    badgeClassName?: string
    errorAlertClassName?: string
}

export const HoverOverlayContents: React.FunctionComponent<HoverOverlayContentsProps> = props => {
    const { hoverOrError, iconClassName, errorAlertClassName, badgeClassName } = props

    const [isRedesignEnabled] = useRedesignToggle()

    if (hoverOrError === 'loading') {
        return (
            <div className="hover-overlay__loader-row">
                <LoadingSpinner className={iconClassName} />
            </div>
        )
    }

    if (isErrorLike(hoverOrError)) {
        return (
            <div
                className={classNames(
                    errorAlertClassName,
                    isRedesignEnabled ? 'hover-overlay__hover-error-redesign' : 'hover-overlay__hover-error'
                )}
            >
                {upperFirst(hoverOrError.message)}
            </div>
        )
    }

    if (hoverOrError === undefined) {
        return null
    }

    if (hoverOrError === null || (hoverOrError.contents.length === 0 && hoverOrError.alerts?.length)) {
        const NoInfoElement = isRedesignEnabled ? 'small' : 'i'

        return (
            // Show some content to give the close button space
            // and communicate to the user we couldn't find a hover.
            <NoInfoElement className="hover-overlay__hover-empty">No hover information available.</NoInfoElement>
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
