import { Icon, NavMenu } from '@sourcegraph/wildcard'
import React from 'react'
import { ConsoleUserData } from '../model'
import { mdiMenuUp, mdiMenuDown } from '@mdi/js'

export const UserMenu: React.FunctionComponent<{ data: ConsoleUserData }> = ({ data }) => (
    <NavMenu
        navTrigger={{
            variant: 'icon',
            triggerContent: {
                trigger: isOpen => (
                    <>
                        {data.user.email}
                        <Icon aria-hidden={true} svgPath={isOpen ? mdiMenuUp : mdiMenuDown} />
                    </>
                ),
            },
        }}
        sections={[
            { headerContent: null, navItems: [{ content: 'Contact support' }] },
            { headerContent: null, navItems: [{ content: 'Sign out' }] },
        ]}
    />
)
