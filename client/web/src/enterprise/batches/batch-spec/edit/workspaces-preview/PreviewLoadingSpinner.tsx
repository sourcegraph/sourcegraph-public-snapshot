import React, { memo } from 'react'

import classNames from 'classnames'

import styles from './PreviewLoadingSpinner.module.scss'

// The SVG is animated with inline <script> tags, so we can't just render it like a normal
// React component and instead load it in with an <object> tag:
// https://www.w3schools.com/tags/tag_object.asp
export const PreviewLoadingSpinner: React.FunctionComponent<
    React.PropsWithChildren<{
        className?: string
    }>
> = memo(({ className }) => {
    const svgSource = `${window.context?.assetsRoot || ''}/img/unoptimized/batchchanges-preview-loading.svg`

    return (
        <object
            type="image/svg+xml"
            data={svgSource}
            className={classNames(className, styles.object)}
            height="50"
            width="50"
        >
            <img src={svgSource} alt="Workspaces preview is loading..." />
        </object>
    )
})
