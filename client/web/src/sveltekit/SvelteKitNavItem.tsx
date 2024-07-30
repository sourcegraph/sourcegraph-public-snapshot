import { FC } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiHelpCircleOutline } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Badge, Icon } from '@sourcegraph/wildcard'

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
                        title="Go to experimental web app"
                        className={styles.toggle}
                    />
                </div>
            </PopoverTrigger>
            <PopoverContent className={styles.popover}>
                <H3>Sourcegraph is getting a refresh!</H3>
                <Text>Try it out early with the toggle above.</Text>
            </PopoverContent>
        </Popover>
    )
}
