import React, { useMemo } from 'react'

import { mdiCheck, mdiInformation } from '@mdi/js'
import classNames from 'classnames'
import MagnifyIcon from 'mdi-react/GearIcon'

import { Icon, Badge, Tooltip } from '@sourcegraph/wildcard'

import { NavDropdown } from '../../nav/NavBar/NavDropdown'

import { useSetupChecklist } from './hooks/useSetupChecklist'

export const ChecklistNavbarItem: React.FC = () => {
    const { data, loading } = useSetupChecklist()
    const notConfiguredCount = useMemo(() => data.filter(item => item.notification).length, [data])

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
            items={data.map(({ notification, id, path, name }) => ({
                content: (
                    <div className="d-flex align-items-center">
                        {notification ? (
                            <Tooltip content={notification.text}>
                                <Icon
                                    svgPath={mdiInformation}
                                    className={classNames(
                                        'mr-1',
                                        notification.type === 'danger' ? 'text-danger' : 'text-warning'
                                    )}
                                    aria-label={notification.text}
                                />
                            </Tooltip>
                        ) : (
                            <Icon svgPath={mdiCheck} aria-label="configured" className="mr-1 text-success" />
                        )}
                        {name}
                    </div>
                ),
                path: notification ? `${path}?setup-checklist=${encodeURIComponent(id)}` : path,
            }))}
            name="feedback"
        />
    )
}
