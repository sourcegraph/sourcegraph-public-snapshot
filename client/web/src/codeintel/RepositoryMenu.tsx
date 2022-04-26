import React from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

export type RepositoryMenuProps = SettingsCascadeProps & {
    repoName: string
    revision: string
    filePath: string
}

// This component is only a stub (hence the null body) that we overwrite in the enterprise
// app. We define this here so we have a stable type to provide on initialization. The OSS
// version simply never renders the code intel repository menu.
export const RepositoryMenu: React.FunctionComponent<RepositoryMenuProps> = () => null
