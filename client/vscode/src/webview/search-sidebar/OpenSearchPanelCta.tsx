import classNames from 'classnames'
import React, { useCallback } from 'react'

import { WebviewPageProps } from '../platform/context'

import styles from './OpenSearchPanelCta.module.scss'

interface OpenSearchPanelCtaProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    className?: string
}

export const OpenSearchPanelCta: React.FunctionComponent<OpenSearchPanelCtaProps> = ({
    sourcegraphVSCodeExtensionAPI,
    className,
}) => {
    const onOpenSearchClick = useCallback(() => sourcegraphVSCodeExtensionAPI.openSearchPanel(), [
        sourcegraphVSCodeExtensionAPI,
    ])

    return (
        <div className={classNames('d-flex flex-column align-items-center justify-content-center', className)}>
            <p className={classNames('my-3', styles.text)}>Search your code and 2M+ open source repositories</p>
            <button
                type="button"
                onClick={onOpenSearchClick}
                className={classNames('btn btn-sm w-100 btn-primary border-0 font-weight-normal', styles.text)}
            >
                Open search panel
            </button>
        </div>
    )
}
