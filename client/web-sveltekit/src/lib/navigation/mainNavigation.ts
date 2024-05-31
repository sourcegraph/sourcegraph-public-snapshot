import type { Page } from '@sveltejs/kit'
import type { ComponentType, SvelteComponent } from 'svelte'
import type { SvelteHTMLElements } from 'svelte/elements'

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
    icon?: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
    /**
     * An optional status to display next to the label. See {@link Status}.
     */
    status?: Status
}

/**
 * A navigation menu is a collection of navigation entries.
 * Currently, it will be rendered as a dropdown in the navigation bar.
 */
export interface NavigationMenu {
    /**
     * The label of the navigation menu.
     */
    label: string
    /**
     * The navigation entries that are part of the menu.
     */
    children: NavigationEntry[]

    /**
     * NavigationMenu item can be rendered as plain link in side navigation mode
     * This fallbackURL will be used to set URL to this link
     */
    href: string

    /**
     * An optional icon to display next to the label.
     */
    icon?: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
    /**
     * A function to determine if current page is part of the menu.
     * This is used to mark the menu as "current" in the UI.
     */
    isCurrent(page: Page): boolean
}

/**
 * A function to determine if a navigation entry is associated with the current page,
 * by means of comparing the entry's href with the current page's URL.
 */
export function isCurrent(entry: NavigationEntry, page: Page): boolean {
    return page.url.pathname === entry.href
}
