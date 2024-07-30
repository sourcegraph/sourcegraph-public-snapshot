import { FC } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiHelpCircleOutline } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Icon } from '@sourcegraph/wildcard'

import { enableSvelteAndReload, canEnableSvelteKit } from './util'

import styles from './SvelteKitNavItem.module.scss'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    return (
        <Popover>
            <PopoverTrigger className={styles.badge}>
                <div className={styles.container}>
                    <Icon svgPath={mdiHelpCircleOutline} />
                    <Text>New, faster UX</Text>
                    <Toggle
                        value={false}
                        onToggle={() => enableSvelteAndReload(client, userID)}
                        title="Enable new, faster UX"
                        className={styles.toggle}
                    />
                </div>
            </PopoverTrigger>
            <PopoverContent className={styles.popover} position="bottomEnd">
                <H3>What’s this "New, faster UX"?</H3>
                <Text>
                    We've been busy at work on a new Code Search experience, built from the ground up for performance,
                    which now available in beta.
                </Text>
                <Text>Simply activate the toggle to get it.</Text>
            </PopoverContent>
        </Popover>
    )
}
