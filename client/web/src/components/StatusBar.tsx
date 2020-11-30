import React, { useEffect, useMemo, useState } from 'react'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { useObservable } from '../../../shared/src/util/useObservable'
import { getStatusBarItems } from '../backend/features'
import classNames from 'classnames'
import iterate from 'iterare'
import { isString } from 'lodash'

interface Props extends StatusBarProps, ExtensionsControllerProps {
    className?: string
}

export interface StatusBarProps {
    onStatusBarVisiblityChange: (visible: boolean) => void
}

export interface StatusBarContainerProps {
    statusBarContainerClassName: string
}

/**
 * Component that fetches and renders status bar items from extensions.
 *
 * Memoized since its props are not likely to change.
 */
export const StatusBar = React.memo<Props>(({ extensionsController, className, onStatusBarVisiblityChange }) => {
    const statusBarItems = useObservable(
        useMemo(() => getStatusBarItems({ extensionsController }), [extensionsController])
    )

    const renderedItems =
        statusBarItems &&
        iterate(statusBarItems)
            .filter(([, item]) => item.visible && isString(item.contentText) && (!item.title || isString(item.title)))
            .map(([id, item]) => (
                <small key={id} className="status-bar__item" data-tooltip={item.hoverMessage}>
                    {item.title && <small className="text-muted">{item.title}: </small>}
                    {item.contentText}
                </small>
            ))
            .toArray()

    const shouldDisplayStatusBar = renderedItems && renderedItems.length > 0

    useEffect(() => {
        onStatusBarVisiblityChange(!!shouldDisplayStatusBar)
    }, [shouldDisplayStatusBar, onStatusBarVisiblityChange])

    return shouldDisplayStatusBar ? (
        <div className={classNames(className, 'status-bar d-flex align-items-center px-2')}>{renderedItems}</div>
    ) : null
})

export function useStatusBar(): StatusBarProps & StatusBarContainerProps {
    const [isStatusBarVisible, setIsStatusBarVisible] = useState(false)

    return {
        statusBarContainerClassName: isStatusBarVisible ? 'pb-4' : '',
        onStatusBarVisiblityChange: setIsStatusBarVisible,
    }
}
