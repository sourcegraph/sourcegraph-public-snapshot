import LinkIcon from 'mdi-react/LinkIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../../schema/settings.schema'

import styles from './SearchSidebarSection.module.scss'

export const getQuickLinks = (settingsCascade: SettingsCascadeProps['settingsCascade']): React.ReactElement[] => {
    const quickLinks = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final.quicklinks) || []

    return quickLinks.map((quickLink, index) => (
        <Link
            key={index}
            to={quickLink.url}
            data-tooltip={quickLink.description}
            className={styles.sidebarSectionListItem}
        >
            <LinkIcon className="icon-inline pr-1" />
            {quickLink.name}
        </Link>
    ))
}
