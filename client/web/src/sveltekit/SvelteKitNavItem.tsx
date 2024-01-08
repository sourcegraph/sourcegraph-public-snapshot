import type { FC } from 'react'

import { useLocation } from 'react-router-dom'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

import { isSvelteKitSupportedURL, reload } from './util'

export const SvelteKitNavItem: FC = () => {
    const location = useLocation()
    const [isSvelteKitToggleEnabled] = useFeatureFlag('enable-sveltekit-toggle')

    if (!isSvelteKitToggleEnabled || !isSvelteKitSupportedURL(location.pathname)) {
        return null
    }

    return (
        <Tooltip content="Go to SvelteKit version">
            <Button variant="icon" onClick={reload}>
                <img src="/.assets/img/svelte-logo-disabled.svg" alt="Svelte Logo" width={20} height={20} />
            </Button>
        </Tooltip>
    )
}
