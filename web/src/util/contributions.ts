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
    /** Whether this route should be handled */
}

/**
 * Used to customize sidebar items
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

export interface ActionButtonDescriptor<C extends object = {}> extends Conditional<C>, WithIcon {
    label: string
    to: string
}
