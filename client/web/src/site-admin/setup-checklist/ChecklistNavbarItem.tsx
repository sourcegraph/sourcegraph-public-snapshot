import React, { useMemo } from 'react'

import { mdiCheck, mdiClose, mdiInformation } from '@mdi/js'
import classNames from 'classnames'
import MagnifyIcon from 'mdi-react/GearIcon'

import { Icon, Badge, Tooltip } from '@sourcegraph/wildcard'

import { NavDropdown } from '../../nav/NavBar/NavDropdown'

import { useSetupChecklist } from './hooks/useSetupChecklist'

export const ChecklistNavbarItem: React.FC = () => {
    const { data, loading } = useSetupChecklist()
    const notConfiguredCount = useMemo(() => data.filter(item => item.needsSetup).length, [data])

    if (loading) {
        return null
    }

    return (
        <NavDropdown
            routeMatch="something-that-does-not-match"
            toggleItem={{
                path: '#',
                icon: MagnifyIcon,
                content: (
                    <div className="d-flex align-items-center">
                        Setup
                        {notConfiguredCount > 0 && (
                            <Badge variant="warning" className="ml-1" pill={true} small={true}>
                                {notConfiguredCount}
                            </Badge>
                        )}
                    </div>
                ),
            }}
            items={data.map(feature => ({
                content: (
                    <div className="d-flex align-items-center">
                        {!feature.needsSetup && (
                            <Icon
                                svgPath={feature.needsSetup ? mdiClose : mdiCheck}
                                aria-label={feature.needsSetup ? 'not configured' : 'configured'}
                                className={classNames('mr-1', feature.needsSetup ? 'text-danger' : 'text-success')}
                            />
                        )}
                        {feature.needsSetup && (
                            <Tooltip content={feature.needsSetup}>
                                <Icon
                                    svgPath={mdiInformation}
                                    // className="ml-1 text-muted"
                                    className={classNames('mr-1 text-danger')}
                                    aria-label={feature.needsSetup}
                                />
                            </Tooltip>
                        )}
                        {feature.name}
                    </div>
                ),
                path: feature?.needsSetup
                    ? `${feature.path}?setup-checklist=${encodeURIComponent(feature.id)}`
                    : feature.path,
            }))}
            name="feedback"
        />
    )
}
