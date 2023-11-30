import React, { useEffect, useMemo, type FC, type PropsWithChildren } from 'react'

import { createRouter, type History, Action, type Location, type Router } from '@remix-run/router'
import type { Navigation } from '@sveltejs/kit'
import { RouterProvider, type RouteObject, UNSAFE_enhanceManualRouteObjects } from 'react-router-dom'

import { RouterLink, setLinkComponent } from '@sourcegraph/wildcard'

import { goto } from '$app/navigation'
import { navigating } from '$app/stores'
import { SettingsProvider, type SettingsCascadeOrError } from '$lib/shared'

import { WildcardThemeContext, type WildcardTheme } from './wildcard'

setLinkComponent(RouterLink)

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

/**
 * Creates a minimal context for rendering React components inside Svelte, including a
 * custom React Router router to integrate with SvelteKit.
 */
export const ReactAdapter: FC<PropsWithChildren<{ route: string; settings: SettingsCascadeOrError }>> = ({
    route,
    children,
    settings,
}) => {
    const router = useMemo(
        () =>
            createSvelteKitRouter([
                {
                    path: route,
                    // React.Suspense seems necessary to render the components without error
                    element: <React.Suspense fallback={true}>{children}</React.Suspense>,
                },
            ]),
        [route, children]
    )

    // Dispose is marked as INTERNAL but without calling it the listeners created by React
    // Router are not removed. It doesn't look like React Router expects to be unmounted
    // during the lifetime of the application.
    useEffect(() => () => router.dispose(), [router])

    return (
        <WildcardThemeContext.Provider value={WILDCARD_THEME}>
            <SettingsProvider settingsCascade={settings}>
                <RouterProvider router={router} />
            </SettingsProvider>
        </WildcardThemeContext.Provider>
    )
}

/**
 * Custom router that synchronizes between the SvelteKit routing and React router.
 */
function createSvelteKitRouter(routes: RouteObject[]): Router {
    return createRouter({
        routes: UNSAFE_enhanceManualRouteObjects(routes),
        history: createSvelteKitHistory(),
    }).initialize()
}

/**
 * Custom history that synchronizes between the SvelteKit routing and React router.
 * This is a "best effort" implementation because the API used here is not very well
 * documented or doesn't seem to be intended for this use case.
 *
 * Caveat: Using the browser back/forward buttons to navigate between hash targets on the
 * same page doesn't seem to scroll the target into view. It's not clear yet why that is.
 */
function createSvelteKitHistory(): History {
    let action: Action = Action.Push

    const history: History = {
        get action() {
            return action
        },

        get location() {
            return createLocation(window.location)
        },

        createURL,

        createHref(to) {
            return typeof to === 'string' ? to : toPath(createURL(to))
        },

        go(delta) {
            window.history.go(delta)
        },

        push(to, state) {
            action = Action.Push
            // Without this, links that are outside of the React component (i.e. Svelte)
            // pointing to a path handle by the React component causes duplicate entries
            // See below for more information.
            // This is safe to do because the browser does the same when clicking on a
            // link that navigates to the curren page.
            if (createURL(to).href !== window.location.href) {
                goto(createURL(to), { state: state ?? undefined })
                    // Make eslint happy
                    .catch(() => {})
            }
        },

        replace(to, state) {
            action = Action.Replace
            goto(createURL(to), { state: state ?? undefined, replaceState: true })
                // Make eslint happy
                .catch(() => {})
        },

        encodeLocation(to) {
            const url = createURL(to)
            return {
                hash: url.hash,
                search: url.search,
                pathname: url.pathname,
            }
        },

        listen(listener) {
            let prevState: Navigation | null = null
            return navigating.subscribe(state => {
                // Events are emitted when navigation *starts*. That means the browser URL hasn't updated yet
                // I don't know whether that's relevant for React Router or not, but in order to make the
                // equality check in the `push` method possible we need to wait to call `listener` until the
                // navigation completed.
                // This is done by storing the emitted value in `prevState` and wait until the next event, which
                // will emit `null`.
                // NOTE: SvelteKit does not emit events when the back/forward buttons are used to navigate between
                // "hashes" on the same page. SvelteKit instead lets the browser handle these natively. However
                // I noticed that at least on Notebook pages this won't scroll the target into view.
                if (!state && prevState) {
                    switch (prevState.type) {
                        case 'popstate': {
                            action = Action.Pop
                            if (prevState.to) {
                                listener({
                                    action,
                                    location: createLocation(prevState.to.url),
                                    delta: prevState.delta ?? 0,
                                })
                            }
                            break
                        }
                        case 'link': {
                            // This is a special case for SvelteKit. In a normal browser context it seems that `listen`
                            // should only handle popstate events. Listening to the SvelteKit 'link' event seems
                            // necessary to properly handle SvelteKit links which point to paths handled by this React
                            // component (it neither works without it nor with the default browser router).
                            // However, React Router doesn't seem to expect that `listener` can be called for "push" events
                            // and will subequently call `history.push`. In order to prevent a double history entry
                            // we are checking whether the target URL is the same as the current URL and do not call
                            // back to SvelteKit again if that's the case.
                            action = Action.Push
                            if (prevState.to) {
                                listener({ action, location: createLocation(prevState.to.url), delta: 1 })
                            }
                            break
                        }
                    }
                }
                prevState = state
            })
        },
    }

    return history
}

function createURL(path: string | { pathname?: string; search?: string; hash?: string }): URL {
    if (typeof path === 'string') {
        return new URL(path, window.location.href)
    }
    const url = new URL(window.location.href)
    if (path.pathname !== undefined) {
        url.pathname = path.pathname
    }
    if (path.search !== undefined) {
        url.search = path.search
    }
    if (path.hash !== undefined) {
        url.hash = path.hash
    }
    return url
}

function createLocation(target: URL | typeof window['location']): Location {
    return {
        pathname: target.pathname,
        search: target.search,
        hash: target.hash,
        state: window.history.state?.usr ?? null,
        key: window.history.state?.key ?? 'default',
    }
}

function toPath(location: { pathname: string; search: string; hash: string }): string {
    return location.pathname + location.search + location.hash
}
