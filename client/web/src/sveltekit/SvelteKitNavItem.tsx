import type { FC } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiFlaskEmptyOutline } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { enableSvelteAndReload, canEnableSvelteKit } from './util'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    return (
        <Tooltip content="Go to experimental web app">
            <Button variant="icon" onClick={() => enableSvelteAndReload(client, userID)}>
                <span className="text-muted">
                    <Icon svgPath={mdiFlaskEmptyOutline} aria-hidden={true} inline={false} />
                </span>
            </Button>
        </Tooltip>
    )
}
