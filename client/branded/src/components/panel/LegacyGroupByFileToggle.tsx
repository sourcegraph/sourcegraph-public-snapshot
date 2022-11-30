import React, { useCallback } from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { updateSettings } from '@sourcegraph/shared/src/api/client/services/settings'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './TabbedPanelContent.module.scss'

interface Props extends PlatformContextProps, SettingsCascadeProps {}

export const LegacyGroupByFileToggle = (props: Props): React.ReactElement | null => {
    const { settingsCascade, platformContext } = props

    const groupByFile = getSettingsValue(settingsCascade, 'panel.locations.groupByFile')

    const label = `${groupByFile ? 'Ungroup' : 'Group'} by file`

    const onClick = useCallback(async () => {
        await updateSettings(platformContext, {
            path: ['panel.locations.groupByFile'],
            value: !groupByFile,
        })
    }, [groupByFile, platformContext])

    return (
        <li className="px-2 mx-2">
            <Tooltip content={null}>
                <Button variant="link" className={styles.actionItem} onClick={onClick}>
                    {label}
                </Button>
            </Tooltip>
        </li>
    )
}

function getSettingsValue(settings: SettingsCascadeOrError, key: string): boolean {
    return !isErrorLike(settings.final) && settings.final !== null && settings.final[key]
}
