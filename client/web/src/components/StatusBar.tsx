import React, { useMemo } from 'react'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { useObservable } from '../../../shared/src/util/useObservable'
import { getStatusBarItems } from '../backend/features'
import classNames from 'classnames'
import iterate from 'iterare'

interface Props extends ExtensionsControllerProps {
    className?: string
}

/**
 * Component that fetches and renders status bar items from extensions
 */
export const StatusBar: React.FunctionComponent<Props> = ({ extensionsController, className }) => {
    const statusBarItems = useObservable(
        useMemo(() => getStatusBarItems({ extensionsController }), [extensionsController])
    )

    if (!statusBarItems) {
        return null
    }

    const renderedItems = iterate(statusBarItems)
        .filter(
            ([, item]) =>
                item.visible &&
                !!item.contentText &&
                typeof item.contentText === 'string' &&
                (typeof item.title !== 'object' || item.title === null)
        )
        .map(([id, item]) => (
            <small key={id} className="status-bar__item">
                {item.title && <small className="text-muted">{item.title}: </small>}
                {item.contentText}
            </small>
        ))
        .toArray()

    return renderedItems.length > 0 ? (
        <div className={classNames(className, 'status-bar d-flex align-items-center px-2')}>{renderedItems}</div>
    ) : null
}
