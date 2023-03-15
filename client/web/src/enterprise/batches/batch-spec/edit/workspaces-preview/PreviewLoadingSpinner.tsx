import React from 'react'

export const PreviewLoadingSpinner: React.FunctionComponent<{
    className?: string
}> = ({ className }) => {
    const svgSource = `${window.context?.assetsRoot || ''}/img/batchchanges-preview-loading.svg`

    return <img src={svgSource} className={className} height="50" width="50" alt="Workspaces preview is loading..." />
}
