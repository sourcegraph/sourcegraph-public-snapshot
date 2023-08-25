import React, { useMemo } from 'react'

import { mdiCheck, mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { sortBy } from 'lodash'
import MagnifyIcon from 'mdi-react/GearIcon'

import { Icon, Badge } from '@sourcegraph/wildcard'

import { NavDropdown } from '../../nav/NavBar/NavDropdown'

import { useSetupChecklist } from './useSetupChecklist'

export const ChecklistNavbarItem: React.FC = () => {
    const { data, loading } = useSetupChecklist()
    const { sortedData, notConfiguredCount } = useMemo(
        () => ({
            sortedData: sortBy(data, item => item.configured),
            notConfiguredCount: data.filter(item => !item.configured).length,
        }),
        [data]
    )

    if (loading) {
        return null
    }

    return (
        <NavDropdown
            routeMatch="something-that-never-matches"
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
            items={sortedData.map(feature => ({
                content: (
                    <div className="d-flex align-items-center">
                        <Icon
                            svgPath={feature.configured ? mdiCheck : mdiClose}
                            aria-label={feature.configured ? 'configured' : 'not configured'}
                            className={classNames('mr-1', feature.configured ? 'text-success' : 'text-danger')}
                        />
                        {feature.name}
                    </div>
                ),
                path: feature.setupURL,
                target: '_blank',
            }))}
            name="feedback"
        />
    )
}
