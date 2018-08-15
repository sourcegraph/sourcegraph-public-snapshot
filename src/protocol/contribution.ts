/** Partial contribution-related client capabilities. */
export interface ContributionClientCapabilities {
    /** The window client capabilities. */
    window?: {
        /** Contribution-related client capabilities. */
        contribution?: {
            /** Whether the client supports dynamic registration of contributions. */
            dynamicRegistration?: boolean
        }
    }
}

/** Partial contribution-related server capabilities. */
export interface ContributionServerCapabilities {
    /** The contributions provided by the server. */
    contributions?: Contributions
}

/**
 * Contributions describes the functionality provided by an extension.
 */
export interface Contributions {
    /** Actions contributed by the extension. */
    actions?: ActionContribution[]

    /** Menu items contributed by the extension. */
    menus?: MenuContributions
}

/**
 * An action contribution describes a command that can be invoked, along with a title, description, icon, etc.
 */
export interface ActionContribution {
    /**
     * The identifier for this action, which must be unique among all contributed actions.
     *
     * Extensions: By convention, this is a dotted string of the form `myExtensionName.myActionName`. It is common
     * to use the same values for `id` and `command` (for the common case where the command has only one action
     * that mentions it).
     */
    id: string

    /**
     * The command that this action invokes. It can refer to a command registered by the same extension or any
     * other extension, or to a builtin command.
     *
     * The builtin command "open" can be used to open a URL (specified as a string in the first element of
     * commandArguments) using the default URL handler on the client, instead of invoking the command on the
     * extension.
     *
     * Extensions: The command must be registered (unless it is a builtin command). Extensions can register
     * commands in the `initialize` response or via `client/registerCapability`.
     *
     * Client: If the command is "open", the client should treat the first element of commandArguments as a URL
     * (string) to open with the default URL handler (instead of sending a request to the extension to execute this
     * command). If the client is running in a web browser, the client should render the action as an HTML <a>
     * element so that it behaves like a link.
     */
    command: string

    /**
     * Optional arguments to pass to the extension when the action is invoked.
     */
    commandArguments?: any[]

    /** The title that succinctly describes what this action does. */
    title?: string

    /**
     * The category that describes the group of related actions of which this action is a member.
     *
     * Clients: When displaying this action's title alongside titles of actions from other groups, the client
     * should display each action as "${category}: ${title}" if the prefix is set.
     */
    category?: string

    /**
     * A longer description of the action taken by this action.
     *
     * Extensions: The description should not be unnecessarily repetitive with the title.
     *
     * Clients: If the description is shown, the title must be shown nearby.
     */
    description?: string

    /**
     * A URL to an icon for this action (data: URIs are OK).
     *
     * Clients: The client should show this icon before the title, proportionally scaling the dimensions as
     * necessary to avoid unduly enlarging the item beyond the dimensions necessary to render the text. The client
     * should assume the icon is square (or roughly square). The client must not display a border around the icon.
     * The client may choose not to display this icon.
     */
    iconURL?: string

    /**
     * A specification of how to display this action as a button on a toolbar. The client is responsible for
     * displaying contributions and defining which parts of its interface are considered to be toolbars. Generally,
     * items on a toolbar are always visible and, compared to items in a dropdown menu or list, are expected to be
     * smaller and to convey information (in addition to performing an action).
     *
     * For example, a "Toggle code coverage" action may prefer to display a summarized status (such as "Coverage:
     * 77%") on a toolbar instead of the full title.
     *
     * Clients: If the label is empty and only an iconURL is set, and the client decides not to display the icon
     * (e.g., because the client is not graphical), then the client may hide the item from the toolbar.
     */
    actionItem?: ActionItem
}

/** A description of how to display a button on a toolbar. */
export interface ActionItem {
    /** The text label for this item. */
    label?: string

    /**
     * A description associated with this action item.
     *
     * Clients: The description should be shown in a tooltip when the user focuses or hovers this toolbar item.
     */
    description?: string

    /**
     * The group in which this item is displayed. This defines the sort order of toolbar items. The group value is
     * an opaque string (it is just compared relative to other toolbar items' group values); there is no
     * specification set of expected or supported values.
     *
     * Clients: On a toolbar, the client should sort toolbar items by (group, command), with toolbar items lacking
     * a group sorting last. The client must not display the group value.
     */
    group?: string

    /**
     * The icon URL for this action (data: URIs are OK).
     *
     * Clients: The client should this icon before the label (if any), proportionally scaling the dimensions as
     * necessary to avoid unduly enlarging the toolbar item beyond the dimensions necessary to show the label.
     * In space-constrained situations, the client should show only the icon and omit the label. The client
     * must not display a border around the icon. The client may choose not to display this icon.
     */
    iconURL?: string

    /**
     * A description of the information represented by the icon.
     *
     * Clients: The client should not display this text directly. Instead, the client should use the
     * accessibility facilities of the client's platform (such as the <img alt> attribute) to make it available
     * to users needing the textual description.
     */
    iconDescription?: string
}

export enum ContributableMenu {
    /** The global command palette. */
    CommandPalette = 'commandPalette',

    /** The global navigation bar in the application. */
    GlobalNav = 'global/nav',

    /** The title bar for the current document. */
    EditorTitle = 'editor/title',

    /** A directory page (including for the root directory of a repository). */
    DirectoryPage = 'directory/page',

    /** The help menu in the application. */
    Help = 'help',
}

/**
 * MenuContributions describes the menu items contributed by an extension.
 */
export interface MenuContributions extends Partial<Record<ContributableMenu, MenuItemContribution[]>> {}

/**
 * MenuItemContribution is a menu item contributed by an extension.
 */
export interface MenuItemContribution {
    /** The action to invoke when selected (== (ActionContribution).id). */
    action: string

    /**
     * An expression that, if given, must evaluate to true (or a truthy value) for this contribution to be
     * displayed. The expression may use values from the context in which the contribution would be displayed.
     */
    when?: string
}
