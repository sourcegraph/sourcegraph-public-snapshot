/**
 * NOTE: should mirror native app's analyticsConstants.ts
 */

export const EventCategories = {
    /**
     * Pageview events
     * Note these are handled specially by the Sourcegraph event logger, and should not be assigned to
     * new events without review
     */
    View: 'View',
    /**
     * Action executed events
     * Note these are handled specially by the Sourcegraph event logger, and should not be assigned to
     * new events without review
     */
    ActionExecuted: 'ActionExecuted',

    /**
     * Events on cross-site navigation elements
     */
    Nav: 'Nav',
    /**
     * Events on static/specific HTML pages
     */
    Pages: 'Pages',
    /**
     * Events on user profile/settings/org pages
     */
    Settings: 'Settings',

    /**
     * Events related to the GitHub/etc authorization process
     */
    Auth: 'Auth',
    /**
     * Events related to the post-auth signup flow
     */
    Onboarding: 'Onboarding',
    /**
     * Events related to marketing, re-engagement or re-targeting, or growth initiatives
     */
    Marketing: 'Marketing',
    /**
     * Events related to sales or online billing
     */
    Billing: 'Billing',

    /**
     * Events related to user state
     */
    UserState: 'UserState',

    /**
     * Launcher events
     */
    Launcher: 'Launcher',
    /**
     * Events on repository pages
     */
    Repository: 'Repository',
    /**
     * Events in VSCode's core editor experience
     */
    Editor: 'Editor',
    /**
     * Events in VSCode's editor sidebar
     */
    EditorSidebar: 'Editor.Sidebar',
    /**
     * Events in (or related to) VSCode's extensions
     */
    Extension: 'Extension',
    /**
     * Events related to any form of search (quick opens, in-repo search, global search, etc)
     */
    Search: 'Search',
    /**
     * Events related to any form of sharing (links, invitations, etc)
     */
    Sharing: 'Sharing',
    /**
     * Events related to providing feedback
     */
    Feedback: 'Feedback',

    /**
     * Events related to VSCode internals
     */
    VSCodeInternal: 'VSCodeInternal',
    /**
     * Events related to VSCode keybindings
     */
    Keys: 'Keys',
    /**
     * Events related to VSCode performance timing/tracking
     */
    Performance: 'Performance',

    /**
     * Events from external applications or pages
     */
    External: 'External',
    /**
     * Other/misc
     */
    Unknown: 'Unknown'
}

export const EventActions = {
    /**
     * Select a result, choice, etc
     */
    Select: 'Select',
    /**
     * Click on a button, link, etc
     */
    Click: 'Click',
    /**
     * Hover over something
     */
    Hover: 'Hover',
    /**
     * Toggle an on/off switch
     */
    Toggle: 'Toggle',

    /**
     * Initiate an action, such as a search
     */
    Initiate: 'Initiate',
    /**
     * Open a window, modal, etc
     */
    Open: 'Open',
    /**
     * Close a window, modal, etc
     */
    Close: 'Close',

    /**
     * Submit a form
     */
    Submit: 'Submit',
    /**
     * Receive a successful response
     */
    Success: 'Success',
    /**
     * Receive an error response
     */
    Error: 'Error',

    Signup: 'Signup',
    Login: 'Login',
    Logout: 'Logout',

    /**
     *  Get redirected from one page or location to another
     */
    Redirect: 'Redirect',

    /**
     * An event that occurs in the background, with no user input
     */
    Passive: 'Passive',
    Unknown: 'Unknown'
}
