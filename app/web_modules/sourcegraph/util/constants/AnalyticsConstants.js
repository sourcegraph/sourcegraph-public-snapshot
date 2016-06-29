// Analytics Constants

// Category constants
// Supply a category name for the group of events you want to track.
// (e.g. 'Video')

// Log when the user performs an action related to authentication. e.g. 'label SignupCompleted'
export const CATEGORY_AUTH = "Auth";
export const CATEGORY_REPOSITORY = "Repository";
export const CATEGORY_HOME = "Home";
export const CATEGORY_TOOLS = "Tools";
export const CATEGORY_DEF_INFO = "DefInfo";
export const CATEGORY_DEF = "Def";
export const CATEGORY_REFERENCES = "References";
export const CATEGORY_PRICING = "Pricing";
export const CATEGORY_GLOBAL_SEARCH = "GlobalSearch";
export const CATEGORY_EXTERNAL = "External";
export const CATEGORY_ENGAGEMENT = "ReEngagement";
export const CATEGORY_UNKNOWN = "Unknown";

// Action constants
// A string that is uniquely paired with each category, and commonly used to define the type of user interaction for the web object.
// (e.g. 'play')
export const ACTION_CLICK = "Click";
export const ACTION_TOGGLE = "Toggle";
export const ACTION_CLOSE = "Close";
export const ACTION_SUCCESS = "Success";
export const ACTION_ERROR = "Error";
export const ACTION_SEARCH = "Search";
export const ACTION_FETCH = "Fetch";
export const ACTION_HOVER = "Hover";
export const ACTION_REDIRECT = "Redirect";

// Label constants
// A string to provide additional dimensions to the event data.
// (e.g 'Fall Campaign')

// Page constants
// Page name constants to provide additional context around where an event was fired from
export const PAGE_HOME = "Home";
export const PAGE_TOOLS = "DashboardTools";
export const PAGE_TOUR = "DashboardTour";
export const PAGE_PRICING = "Pricing";

// Location on page constants
// Page location constants are used to provide additional context around where on a page an event was fired from
export const PAGE_LOCATION_GLOBAL_NAV = "GlobalNav";

// Integrations constants
// Names for Sourcegraph integrations
export const INTEGRATION_EDITOR_SUBLIME = "Sublime";
export const INTEGRATION_EDITOR_EMACS = "Emacs";
export const INTEGRATION_EDITOR_VIM = "VIM";
