import React from 'react'

import { ExtensionViewsDirectorySection, ExtensionViewsDirectorySectionProps } from './ExtensionViewsDirectorySection'
import { ExtensionViewsHomepageSection, ExtensionViewsHomepageSectionProps } from './ExtensionViewsHomepageSection'

export type ExtensionViewsSectionProps = ExtensionViewsDirectorySectionProps | ExtensionViewsHomepageSectionProps

/**
 * Renders section extension views section based on where prop. This component is used only for
 * Enterprise version. See `./src/insight/sections` components for the OSS version.
 */
export const ExtensionViewsSection: React.FunctionComponent<ExtensionViewsSectionProps> = props => {
    const { where } = props

    switch (where) {
        case 'directory':
            return <ExtensionViewsDirectorySection {...props} />
        case 'homepage':
            return <ExtensionViewsHomepageSection {...props} />

        default:
            return null
    }
}
