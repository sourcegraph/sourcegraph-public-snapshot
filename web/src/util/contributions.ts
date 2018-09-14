import { RouteComponentProps } from 'react-router'

interface Conditional<C extends object> {
    /** Optional condition under which this item should be used */
    condition?: (context: C) => boolean
}

interface WithIcon {
    icon?: React.ComponentType<{ className?: string }>
}

/**
 * Configuration for a route.
 *
 * @template C Context information that is passed to `render` and `condition`
 */
export interface RouteDescriptor<C extends object = {}> extends Conditional<C> {
    /** Path of this route (appended to the current match) */
    path: string
    exact?: boolean
    render: ((props: C & RouteComponentProps<any>) => React.ReactNode)
}

/**
 * Used to customize sidebar items.
 * The difference between this and an action button is that nav items get highlighted if their `to` route matches.
 *
 * @template C Context information that is made available to determine whether the item should be shown (different for each sidebar)
 */
export interface NavItemDescriptor<C extends object = {}> extends Conditional<C> {
    /** The text of the item */
    label: string

    /** The link destination (appended to the current match) */
    to: string

    /** Whether highlighting the item should only be done if `to` matches exactly */
    exact?: boolean
}

export interface NavItemWithIconDescriptor<C extends object = {}> extends NavItemDescriptor<C>, WithIcon {}

/**
 * A descriptor for an action button that should appear somewhere in the UI.
 *
 * @template C Context information that is made available to determine whether the item should be shown and the link destination
 */
export interface ActionButtonDescriptor<C extends object = {}> extends Conditional<C>, WithIcon {
    /** Label for for the button  */
    label: string

    /** Optional tooltip for the button (if set, should include more information than the label) */
    tooltip?: string

    /** Function to return the destination link for the button */
    to: (context: C) => string
}
