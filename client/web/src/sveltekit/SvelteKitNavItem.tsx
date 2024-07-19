import { FC } from 'react'

import { useApolloClient } from '@apollo/client'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Badge } from '@sourcegraph/wildcard'

import { enableSvelteAndReload, canEnableSvelteKit } from './util'

import styles from './SvelteKitNavItem.module.scss'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    return (
        <div className={styles.container}>
            <Toggle
                value={false}
                onToggle={() => enableSvelteAndReload(client, userID)}
                title="Go to experimental web app"
                className={styles.toggle}
            />
            <Popover>
                <PopoverTrigger className={styles.badge}>
                    <Badge variant="warning">Try the new experience</Badge>
                </PopoverTrigger>
                <PopoverContent className={styles.popover}>
                    <H3>Sourcegraph is getting a refresh!</H3>
                    <Text>Try it out early with the toggle above.</Text>
                </PopoverContent>
            </Popover>
        </div>
    )
}
