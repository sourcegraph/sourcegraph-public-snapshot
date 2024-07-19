import { type useState, FC } from 'react'

import { useApolloClient } from '@apollo/client'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Badge } from '@sourcegraph/wildcard/src/components/Badge'
import { Popover, PopoverTrigger, PopoverContent } from '@sourcegraph/wildcard/src/components/Popover'

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
                title={'Go to experimental web app'}
                className={styles.toggle}
            />
            <Popover>
                <PopoverTrigger className={styles.badge}>
                    <Badge variant={'warning'}>Try the new experience</Badge>
                </PopoverTrigger>
                <PopoverContent className={styles.popover}>
                    <h3>Sourcegraph is getting a refresh!</h3>
                    <p>Try it out early with the toggle above.</p>
                </PopoverContent>
            </Popover>
        </div>
    )
}
