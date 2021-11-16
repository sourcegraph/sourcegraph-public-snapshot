import React, { memo } from 'react'

// The SVG is animated with inline <script> tags, so we can't just render it like a normal
// React component and instead load it in with an <object> tag:
// https://www.w3schools.com/tags/tag_object.asp
export const PreviewLoadingSpinner: React.FunctionComponent<{
    className?: string
}> = memo(({ className }) => {
    const svgSource = `${window.context?.assetsRoot || ''}/img/batchchanges-preview-loading.svg`

    return (
        <object type="image/svg+xml" data={svgSource} className={className} height="50" width="50">
            <img src={svgSource} alt="Workspaces preview is loading..." />
        </object>
    )
})
