import type { FC } from 'react'

import { mdiFlaskEmptyOutline } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

import { isSvelteKitSupportedURL, reload } from './util'

function useIsSvelteKitToggleEnabled(): boolean {
    const [isSvelteKitToggleEnabled] = useFeatureFlag('enable-sveltekit-toggle')
    const [isExperimentalWebAppToggleEnabled] = useFeatureFlag('web-next-toggle')
    return isSvelteKitToggleEnabled || isExperimentalWebAppToggleEnabled
}

export const SvelteKitNavItem: FC = () => {
    const location = useLocation()
    const isEnabled = useIsSvelteKitToggleEnabled()

    if (!isEnabled || !isSvelteKitSupportedURL(location.pathname)) {
        return null
    }

    return (
        <Tooltip content="Go to experimental web app">
            <Button variant="icon" onClick={reload}>
                <span className="text-muted">
                    <Icon svgPath={mdiFlaskEmptyOutline} aria-hidden={true} inline={false} />
                </span>
            </Button>
        </Tooltip>
    )
}
