import { KeyPath } from './configuration'

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
     * Extensions: The command must be registered (unless it is a builtin command). Extensions can register
     * commands in the `initialize` response or via `client/registerCapability`.
     *
     * ## Builtin client commands
     *
     * Clients: All clients must handle the following commands as specified.
     *
     * ### `open` {@link ActionContributionClientCommandOpen}
     *
     * The builtin command `open` causes the client to open a URL (specified as a string in the first element of
     * commandArguments) using the default URL handler, instead of invoking the command on the extension.
     *
     * Clients: The client should treat the first element of commandArguments as a URL (string) to open with the
     * default URL handler (instead of sending a request to the extension to execute this command). If the client
     * is running in a web browser, the client should render the action as an HTML <a> element so that it behaves
     * like a link.
     *
     * ### `updateConfiguration` {@link ActionContributionClientCommandUpdateConfiguration}
     *
     * The builtin command `updateConfiguration` causes the client to apply an update to the configuration settings.
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

/**
 * Narrows the type of {@link ActionContribution} for actions that invoke the `open` client command.
 */
export interface ActionContributionClientCommandOpen extends ActionContribution {
    command: 'open'

    /**
     * The arguments for the `open` client command. The first array element is a URL, which is opened by the client
     * using the default URL handler.
     */
    commandArguments: [string]
}

/**
 * Narrows the type of {@link ActionContribution} for actions that invoke the `updateConfiguration` client command.
 */
export interface ActionContributionClientCommandUpdateConfiguration extends ActionContribution {
    command: 'updateConfiguration'

    /**
     * The arguments for the `updateConfiguration` client command:
     *
     * 1. The key path of the value (in the configuration settings) to update
     * 2. The value to insert at the key path
     * 3. Optional: reserved for future use (must always be `null`)
     * 4. Optional: 'json' if the client should parse the 2nd argument using JSON.parse before inserting the value
     */
    commandArguments: [KeyPath, any] | [KeyPath, string, null, 'json']
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
    /**
     * The action to invoke when the item is selected. The value refers to a {@link ActionContribution#id} value.
     */
    action: string

    /**
     * An alternative action to invoke when the item is selected while pressing the Option/Alt/Meta/Ctrl/Cmd keys
     * or using the middle mouse button. The value refers to a {@link ActionContribution#id} value.
     */
    alt?: string

    /**
     * An expression that, if given, must evaluate to true (or a truthy value) for this contribution to be
     * displayed. The expression may use values from the context in which the contribution would be displayed.
     */
    when?: string

    /**
     * The group in which this item is displayed. This defines the sort order of menu items. The group value is an
     * opaque string (it is just compared relative to other items' group values); there is no specification set of
     * expected or supported values.
     *
     * Clients: On a toolbar, the client should sort toolbar items by (group, action), with toolbar items lacking a
     * group sorting last. The client must not display the group value.
     */
    group?: string
}

/** The containers to which an extension can contribute views. */
export enum ContributableViewContainer {
    /**
     * A view that is displayed in the panel for a window.
     *
     * Clients: The client should render this as a resizable panel in a window, with multiple tabs to switch
     * between different panel views.
     */
    Panel = 'window/panel',
}
