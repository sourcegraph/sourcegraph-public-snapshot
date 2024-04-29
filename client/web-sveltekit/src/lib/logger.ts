// We want to limit the number of imported modules as much as possible

import { onMount } from 'svelte'

import { PUBLIC_ENABLE_EVENT_LOGGER } from '$env/static/public'

/**
 * Can only be called during component initialization. It logs a view event when
 * the component is mounted (and event logging is enabled).
 */
export function logViewEvent(eventName: string, eventProperties?: any, publicArgument?: any): void {
    if (PUBLIC_ENABLE_EVENT_LOGGER) {
        onMount(() => {
            // TODO: Implement event logging
            console.log('logViewEvent', eventName, eventProperties, publicArgument)
        })
    }
}
