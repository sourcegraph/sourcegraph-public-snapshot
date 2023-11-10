import type React from 'react'

interface Conditional<C extends object> {
    /** Optional condition under which this item should be used */
    readonly condition?: (context: C) => boolean
}

interface WithIcon {
    readonly icon?: React.ComponentType<{ className?: string }>
}

/**
 * Configuration for a component.
 *
 * @template C Context information that is passed to `render` and `condition`
 */
export interface ComponentDescriptor<C extends object = {}> extends Conditional<C> {
    readonly render: (props: C) => React.ReactNode
}

/**
 * Configuration for a react-router 6 route.
 *
 * @template C Context information that is passed to `render` and `condition`
 */
export interface RouteV6Descriptor<C extends object = {}> extends Conditional<C> {
    readonly path: string
    readonly render: (props: C) => React.ReactNode
    readonly exact?: boolean
}

export interface NavGroupDescriptor<C extends object = {}> extends Conditional<C> {
    readonly header?: {
        readonly label: string
        readonly icon?: React.ComponentType<{ className?: string }>
        readonly source?: 'server' | 'client'
    }
    readonly items: readonly NavItemDescriptor<C>[]
}

/**
 * Used to customize sidebar items.
 * The difference between this and an action button is that nav items get highlighted if their `to` route matches.
 *
 * @template C Context information that is made available to determine whether the item should be shown (different for each sidebar)
 */
export interface NavItemDescriptor<C extends object = {}> extends Conditional<C> {
    /** The text of the item */
    readonly label: string

    /** The link destination (appended to the current match) */
    readonly to: string

    /** Whether highlighting the item should only be done if `to` matches exactly */
    readonly exact?: boolean

    /** The link source to determine the render strategy (react-dom link component or <a>) */
    readonly source?: 'server' | 'client'

    /** The text of the item calculated dynamically passing the context*/
    readonly dynamicLabel?: (context: C) => React.ReactNode
}

export interface NavItemWithIconDescriptor<C extends object = {}> extends NavItemDescriptor<C>, WithIcon {}

/**
 * A descriptor for an action button that should appear somewhere in the UI.
 *
 * @template C Context information that is made available to determine whether the item should be shown and the link destination
 */
export interface ActionButtonDescriptor<C extends object = {}> extends Conditional<C>, WithIcon {
    /** Label for for the button  */
    readonly label: string

    /** Optional tooltip for the button (if set, should include more information than the label) */
    readonly tooltip?: string

    /** Function to return the destination link for the button */
    readonly to: (context: C) => string
}
