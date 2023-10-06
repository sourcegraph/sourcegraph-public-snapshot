import type { FC } from 'react'

import { mdiAlertOctagon } from '@mdi/js'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { NavAction } from '../nav'
import { toggleDevSettingsDialog, useOverrideCounter } from '../stores'

export const DeveloperSettingsGlobalNavItem: FC<{}> = () => {
    const counter = useOverrideCounter()
    const hasOverrides = counter.featureFlags + counter.temporarySettings > 0

    return (
        <NavAction>
            <span>
                <Button variant="link" onClick={() => toggleDevSettingsDialog(true)}>
                    Developer Settings
                    {hasOverrides && (
                        <Tooltip
                            content={`You have ${counter.featureFlags} local feature flag and ${counter.temporarySettings} temporary settings overrides.`}
                        >
                            <Icon
                                className="ml-1"
                                style={{ color: 'var(--orange)' }}
                                svgPath={mdiAlertOctagon}
                                aria-hidden={true}
                            />
                        </Tooltip>
                    )}
                </Button>
            </span>
        </NavAction>
    )
}
