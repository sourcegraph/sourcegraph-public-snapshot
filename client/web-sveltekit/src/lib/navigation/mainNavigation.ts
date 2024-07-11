import type { Page } from '@sveltejs/kit'

import type { IconComponent } from '$lib/Icon.svelte'

/**
 * Indicates to the UI to show a status badge next to the navigation entry.
 */
export enum Status {
    BETA = 1,
}

/**
 * A navigation entry is a single item in the navigation menu
 * It is a link to a specific page in the application.
 */
export interface NavigationEntry {
    /**
     * The label of the navigation entry.
     */
    label: string
    /**
     * The target URL to navigate to.
     */
    href: string
    /**
     * An optional icon to display next to the label.
     */
    icon?: IconComponent
    /**
     * An optional status to display next to the label. See {@link Status}.
     */
    status?: Status
}

/**
 * A navigation menu is a collection of navigation entries.
 * Currently, it will be rendered as a dropdown in the navigation bar.
 */
export interface NavigationMenu<T extends NavigationEntry = NavigationEntry> {
    /**
     * The label of the navigation menu.
     */
    label: string
    /**
     * The navigation entries that are part of the menu.
     */
    children: T[]

    /**
     * Target URL to navigate to when the menu is clicked.
     */
    href: string

    /**
     * An optional icon to display next to the label.
     */
    icon?: IconComponent
}

/**
 * A function to determine if a navigation entry is associated with the current page,
 * by means of comparing the entry's href with the current page's URL.
 */
export function isCurrent(entry: NavigationEntry, page: Page): boolean {
    return page.url.pathname === entry.href
}

export function isNavigationMenu(entry: NavigationEntry | NavigationMenu): entry is NavigationMenu {
    return entry && 'children' in entry && entry.children.length > 0
}
