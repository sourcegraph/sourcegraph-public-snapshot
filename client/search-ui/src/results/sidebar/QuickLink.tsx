import LinkIcon from 'mdi-react/LinkIcon'
import React from 'react'

import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/client-api'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { Link } from '@sourcegraph/wildcard'

import styles from './SearchSidebarSection.module.scss'

export const getQuickLinks = (settingsCascade: SettingsCascadeProps['settingsCascade']): React.ReactElement[] => {
    const quickLinks = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final.quicklinks) || []

    return quickLinks.map((quickLink, index) => (
        <Link
            // Can't guarantee that URL, name, or description are unique, so use index as key.
            // This is safe since this list will only be updated when settings change.
            // eslint-disable-next-line react/no-array-index-key
            key={index}
            to={quickLink.url}
            data-tooltip={quickLink.description}
            data-placement="right"
            className={styles.sidebarSectionListItem}
        >
            <LinkIcon className="icon-inline pr-1 flex-shrink-0" />
            {quickLink.name}
        </Link>
    ))
}
