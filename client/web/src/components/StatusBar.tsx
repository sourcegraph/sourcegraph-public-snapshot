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

    return (
        <div className={classNames(className)}>
            {iterate(statusBarItems)
                .filter(([, item]) => item.visible && !!item.contentText)
                .map(([id, item]) => <p key={id}>{item.contentText}</p>)
                .toArray()}
        </div>
    )
}
