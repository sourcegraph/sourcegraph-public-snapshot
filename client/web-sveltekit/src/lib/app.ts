import { onDestroy } from 'svelte'

import { beforeNavigate } from '$app/navigation'
import { page } from '$app/stores'

const scrollCache: Map<string, number> = new Map()

/**
 * Stores and restores scroll position. Needs to be called at component
 * initialization time.
 *
 * `setter` is called when the component is instantiated. The value passed is
 * the previously stored scroll position, if any. This value can then be used to
 * set the target elements scroll position when it becomes available.
 *
 * `getter` is called when the current page is being navigated away from. It
 * should to return the scroll position of the target element.
 *
 * Example:
 *
 *  let scrollTop: number|undefined
 *  let element: HTMLElement|undefined
 *
 *  preserveScrollPosition(
 *      position => (scrollTop = position ?? 0),
 *      () => resultContainer?.scrollTop
 *  )
 *  $: if (element) {
 *      element.scrollTop = scrollTop ?? 0
 *  }
 *  ...
 *  <div bind:this={element} />
 */
export function preserveScrollPosition(
    setter: (position: number | undefined) => void,
    getter: () => number | undefined
): void {
    onDestroy(
        page.subscribe($page => {
            setter(scrollCache.get($page.url.toString()))
        })
    )

    beforeNavigate(({ from }) => {
        if (from) {
            const position = getter()
            if (position) {
                scrollCache.set(from?.url.toString(), position)
            }
        }
    })
}
