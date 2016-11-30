import { EventLogger } from "sourcegraph/util/EventLogger";

// Analytics Constants

// Category constants
// Supply a category name for the group of events you want to track.

export const EventCategories = {
	// Home pages
	Nav: "Nav",
	Home: "Home",
	Dashboard: "Dashboard",
	Pricing: "Pricing",
	DocsPage: "DocsPage",
	Orgs: "Orgs",
	Settings: "Settings",
	Tools: "Tools",

	// Auth and marketing
	Auth: "Auth",
	Onboarding: "Onboarding",
	ReEngagement: "ReEngagement",
	Toast: "Toast",

	// Application pages
	Repository: "Repository",
	LandingDefInfo: "LandingDefInfo",
	QuickOpen: "QuickOpen",
	CodeView: "CodeView",

	// Misc other
	External: "External",
	Unknown: "Unknown",

	// Deprecated
	Def: "Def",
	References: "References",
	File: "File",
	GlobalSearch: "GlobalSearch",
};

// Action constants
// A string that is uniquely paired with each category, and commonly used to define the type of user interaction for the web object.
// (e.g. 'play')
export const EventActions = {
	Click: "Click",
	Toggle: "Toggle",
	Close: "Close",
	Success: "Success",
	Error: "Error",
	Search: "Search",
	Fetch: "Fetch",
	Hover: "Hover",
	Redirect: "Redirect",
	Signup: "Signup",
	Login: "Login",
	Logout: "Logout",
};

// Label constants
// A string to provide additional dimensions to the event data.
// (e.g 'Fall Campaign')

// Page constants
// Page name constants to provide additional context around where an event was fired from
export const PAGE_HOME = "Home";
export const PAGE_DASHBOARD = "Dashboard";
export const PAGE_TOOLS = "DashboardTools";
export const PAGE_PRICING = "Pricing";

// Location on page constants
// Page location constants are used to provide additional context around where on a page an event was fired from
export const PAGE_LOCATION_GLOBAL_NAV = "GlobalNav";
export const PAGE_LOCATION_DASHBOARD = "Dashboard";

// Integrations constants
// Names for Sourcegraph integrations
export const INTEGRATION_EDITOR_SUBLIME = "Sublime";
export const INTEGRATION_EDITOR_EMACS = "Emacs";
export const INTEGRATION_EDITOR_VIM = "VIM";

export class LoggableEvent {
	label: string;
	category: string;
	action: string;
	permittedProps?: any;

	constructor(label: string, category: string, action: string, permittedProps?: any) {
		this.label = label;
		this.category = category;
		this.action = action;
		this.permittedProps = permittedProps;
	}

	logEvent(props?: any): void {
		EventLogger.logEventForCategory(this, props);
	}
}
export class NonInteractionLoggableEvent extends LoggableEvent {
	logEvent(props?: any): void {
		EventLogger.logNonInteractionEventForCategory(this, props);
	}
}

// List of all possible events, paired with their unique labels, categories, and user actions
// Calls to EventLogger.logEventForCategory() shouldn't require category/action params, given each event can have only a single one of each of those
// TODO: Add list of permissable properties to each event to limit proliferation
export const Events = {
	// NOTE: this is only a repository for sourcegraph.com events; others include:
	// - Chrome extension 
	//      - Listed in app/analytics/EventLogger.js in the browser-ext folder  

	// Auth events
	ContactIntercom_Clicked: new LoggableEvent("ClickContactIntercom", EventCategories.Auth, EventActions.Click),
	OAuth2FlowGitHub_Initiated: new LoggableEvent("InitiateGitHubOAuth2Flow", EventCategories.Auth, EventActions.Click),
	OAuth2FlowGitHub_Completed: new LoggableEvent("CompletedGitHubOAuth2Flow", EventCategories.Auth, EventActions.Login),
	OAuth2FlowGCP_Initiated: new LoggableEvent("InitiateGCPOAuth2Flow", EventCategories.Auth, EventActions.Click),
	OAuth2FlowGCP_Completed: new LoggableEvent("CompletedGCPOAuth2Flow", EventCategories.Auth, EventActions.Login),
	Signup_Completed: new LoggableEvent("SignupCompleted", EventCategories.Auth, EventActions.Signup),
	Logout_Clicked: new LoggableEvent("LogoutClicked", EventCategories.Auth, EventActions.Logout),

	// Modals
	JoinModal_Initiated: new LoggableEvent("ShowSignUpModal", EventCategories.Auth, EventActions.Toggle),
	LoginModal_Initiated: new LoggableEvent("ShowLoginModal", EventCategories.Auth, EventActions.Toggle),
	ToolsModal_Initiated: new LoggableEvent("ClickToolsandIntegrations", EventCategories.Auth, EventActions.Toggle),
	BetaModal_Initiated: new LoggableEvent("ClickJoinBeta", EventCategories.Auth, EventActions.Toggle),
	JoinModal_Dismissed: new LoggableEvent("DismissJoinModal", EventCategories.Auth, EventActions.Close),
	LoginModal_Dismissed: new LoggableEvent("DismissLoginModal", EventCategories.Auth, EventActions.Close),
	ToolsModal_Dismissed: new LoggableEvent("DismissToolsandIntegrationsModal", EventCategories.Auth, EventActions.Close),
	BetaModal_Dismissed: new LoggableEvent("DismissBetaModal", EventCategories.Auth, EventActions.Close),

	// Toast events
	ToastChromeCTA_Clicked: new LoggableEvent("ChromeToastCTAClicked", EventCategories.Toast, EventActions.Click),
	ToastChrome_Closed: new LoggableEvent("ChromeToastCloseClicked", EventCategories.Toast, EventActions.Close),

	// Repo events
	Repository_Clicked: new LoggableEvent("RepoClicked", EventCategories.Repository, EventActions.Click),
	Repository_Added: new LoggableEvent("AddRepo", EventCategories.Repository, EventActions.Success),
	RepositoryAuthedLanguagesGitHub_Fetched: new LoggableEvent("AuthedLanguagesGitHubFetched", EventCategories.Repository, EventActions.Fetch),
	RepositoryAuthedReposGitHub_Fetched: new LoggableEvent("AuthedReposGitHubFetched", EventCategories.Repository, EventActions.Fetch),

	// Onboarding
	OnboardingRefsCoachCTA_Clicked: new LoggableEvent("ReferencesCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingJ2DCoachCTA_Clicked: new LoggableEvent("JumpToDefCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingSearchCoachCTA_Clicked: new LoggableEvent("SearchCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingTour_Completed: new LoggableEvent("OnboardingTourCompleted", EventCategories.Onboarding, EventActions.Success),
	OnboardingTour_Dismissed: new LoggableEvent("DismissTourCTAClicked", EventCategories.Onboarding, EventActions.Close),

	ChromeExtension_Installed: new LoggableEvent("ChromeExtensionInstalled", EventCategories.Onboarding, EventActions.Success),
	ChromeExtensionInstall_Failed: new LoggableEvent("ChromeExtensionInstallFailed", EventCategories.Onboarding, EventActions.Error),
	ChromeExtensionCTA_Clicked: new LoggableEvent("ChromeExtensionCTAClicked", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionInstall_Started: new LoggableEvent("ChromeExtensionInstallStarted", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionStore_Redirected: new LoggableEvent("ChromeExtensionStoreRedirect", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionSkipCTA_Clicked: new LoggableEvent("SkipChromeExtensionCTAClicked", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionUnsupportedBrowser_Failed: new LoggableEvent("BrowserDoesNotSupportChromeExtension", EventCategories.Onboarding, EventActions.Error),
	ChromeExtensionStep_Completed: new LoggableEvent("ChromeExtensionStepCompleted", EventCategories.Onboarding, EventActions.Success),

	AuthGitHubStep_Completed: new LoggableEvent("GitHubStepCompleted", EventCategories.Onboarding, EventActions.Success),
	AuthGCPStep_Completed: new LoggableEvent("GCPStepCompleted", EventCategories.Onboarding, EventActions.Success),
	PrivateAuthGitHub_Skipped: new LoggableEvent("SkipGitHubPrivateAuth", EventCategories.Onboarding, EventActions.Click),
	PrivateAuthGCP_Skipped: new LoggableEvent("SkipGCPPrivateAuth", EventCategories.Onboarding, EventActions.Click),

	// ReEngagement
	BetaSubscription_Completed: new LoggableEvent("BetaSubscriptionCompleted", EventCategories.ReEngagement, EventActions.Success),

	// Code view
	CodeContextMenu_Initiated: new LoggableEvent("CodeContextMenuClicked", EventCategories.CodeView, EventActions.Click),
	CodeReferences_Viewed: new LoggableEvent("ClickedViewReferences", EventCategories.CodeView, EventActions.Click),
	CodeExternalReferences_Viewed: new LoggableEvent("ClickedViewExternalReferences", EventCategories.CodeView, EventActions.Click),
	CodeToken_Hovered: new LoggableEvent("Hovering", EventCategories.CodeView, EventActions.Hover),
	CodeToken_Clicked: new LoggableEvent("BlobTokenClicked", EventCategories.CodeView, EventActions.Click),
	OpenInCodeHost_Clicked: new LoggableEvent("OpenInCodeHostClicked", EventCategories.CodeView, EventActions.Click),
	FileTree_Navigated: new LoggableEvent("FileTreeActivated", EventCategories.CodeView, EventActions.Click),

	// Quick open/search
	QuickopenItem_Selected: new LoggableEvent("QuickOpenItemSelected", EventCategories.QuickOpen, EventActions.Click),
	Quickopen_Initiated: new LoggableEvent("QuickOpenInitiated", EventCategories.QuickOpen, EventActions.Toggle),
	Quickopen_Dismissed: new LoggableEvent("QuickOpenDismissed", EventCategories.QuickOpen, EventActions.Close),

	// Def info - implemented manually in static_event_logger.js 
	// 	TODO(dadlerj): delete these, or integrate this file with that logger
	// DefInfoDefLink_Clicked:                  new LoggableEvent("DefInfoViewDefLinkClicked", EventCategories.LandingDefInfo, EventActions.Click),
	// DefInfoFileLink_Clicked:                 new LoggableEvent("DefInfoViewFileLinkClicked", EventCategories.LandingDefInfo, EventActions.Click),
	// DefInfoRefSnippedLink_Clicked:           new LoggableEvent("DefInfoRefSnippetLinkClicked", EventCategories.LandingDefInfo, EventActions.Click),
	// DefInfoRefRepoLink_Clicked:              new LoggableEvent("DefInfoRefRepoLinkClicked", EventCategories.LandingDefInfo, EventActions.Click),
	// DefInfoRefFileLink_Clicked:              new LoggableEvent("DefInfoRefFileLinkClicked", EventCategories.LandingDefInfo, EventActions.Click),

	// Orgs
	Org_Selected: new LoggableEvent("SelectedOrg", EventCategories.Orgs, EventActions.Click),
	OrgUser_Invited: new LoggableEvent("InviteUser", EventCategories.Orgs, EventActions.Success),
	OrgManualInviteModal_Initiated: new LoggableEvent("ToggleManualInviteModal", EventCategories.Orgs, EventActions.Toggle),
	OrgManualInviteModal_Dismissed: new LoggableEvent("DismissManualInviteModal", EventCategories.Orgs, EventActions.Close),
	OrgEmailInvite_Clicked: new LoggableEvent("EmailInviteClicked", EventCategories.Orgs, EventActions.Click),
	AuthedOrgsGitHub_Fetched: new LoggableEvent("AuthedOrgsGitHubFetched", EventCategories.Orgs, EventActions.Fetch),
	AuthedOrgMembersGitHub_Fetched: new LoggableEvent("AuthedOrgMembersGitHubFetched", EventCategories.Orgs, EventActions.Fetch),

	// Settings (org and repo views)
	SettingsRepoView_Toggled: new LoggableEvent("ToggleRepoView", EventCategories.Settings, EventActions.Toggle),
	SettingsOrgView_Toggled: new LoggableEvent("ToggleOrgView", EventCategories.Settings, EventActions.Toggle),

	// Static pages
	DocsContactSupportCTA_Clicked: new LoggableEvent("clickedContactSupportFromDocs", EventCategories.DocsPage, EventActions.Click),
	DocsInstallExtensionCTA_Clicked: new LoggableEvent("clickedInstallBrowserExtFromDocs", EventCategories.DocsPage, EventActions.Click),
	DocsAuthPrivateCTA_Clicked: new LoggableEvent("clickedAuthPrivateReposFromDocs", EventCategories.DocsPage, EventActions.Click),
	ToolsModalDownloadCTA_Clicked: new LoggableEvent("ToolsModalDownloadCTAClicked", EventCategories.Tools, EventActions.Click),
	PricingCTA_Clicked: new LoggableEvent("ClickPricingCTA", EventCategories.Pricing, EventActions.Click),
	DashboardRepo_Clicked: new LoggableEvent("DashboardRepoClicked", EventCategories.Dashboard, EventActions.Click),
	HomeCarousel_Clicked: new LoggableEvent("HomeCarouselClicked", EventCategories.Home, EventActions.Click),
	JobsCTA_Clicked: new LoggableEvent("JobsCTAClicked", EventCategories.Nav, EventActions.Click),

	// Non-Interaction Events
	// Events that we wish to track, but do not wish to impact bounce rate on our site for Google analytics.
	// See EventLogger.logNonInteractionEventForCategory() for more information
	ViewRepoMain_Failed: new NonInteractionLoggableEvent("ViewRepoMainError", EventCategories.Repository, EventActions.Error),
};

export function getModalDismissedEventObject(modalName: string): LoggableEvent {
	const dismissModalsMap = {
		"beta": Events.BetaModal_Dismissed,
		"menuBeta": Events.BetaModal_Dismissed,
		"menuIntegrations": Events.ToolsModal_Dismissed,
		"join": Events.JoinModal_Dismissed,
		"login": Events.LoginModal_Dismissed,
		"orgInvite": Events.OrgManualInviteModal_Dismissed,
	};
	return (modalName && modalName in dismissModalsMap) ? dismissModalsMap[modalName] : null;
}
