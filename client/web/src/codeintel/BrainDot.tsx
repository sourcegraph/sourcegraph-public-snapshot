import React from 'react'

export interface BrainDotProps {
    repoName: string
    commit: string
    path?: string
    visibleIndexID?: string
    setVisibleIndexID: (id?: string) => void
}

// This component is only a stub (hence the null body) that we overwrite in the enterprise
// app. We define this here so we have a stable type to provide on initialization. The OSS
// version simply never renders the code intel repository menu.
export const BrainDot: React.FunctionComponent<BrainDotProps> = () => null
