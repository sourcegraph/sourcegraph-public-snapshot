import { EventLogger } from "sourcegraph/tracking/EventLogger";

// Analytics Constants

// Category constants
// Supply a category name for the group of events you want to track.

export const EventCategories = {
	// Home pagess
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
	GTM: "GTM",
	Billing: "Billing",

	// Application pages
	Repository: "Repository",
	LandingDefInfo: "LandingDefInfo",
	QuickOpen: "QuickOpen",
	CodeView: "CodeView",

	// Misc other
	External: "External",
	Unknown: "Unknown",
	ShortcutMenu: "ShortcutMenu",

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
export const PAGE_LOCATION_DASHBOARD_SIDEBAR = "DashboardSidebar";

// Integrations constants
// Names for Sourcegraph integrations
export const INTEGRATION_EDITOR_SUBLIME = "Sublime";
export const INTEGRATION_EDITOR_EMACS = "Emacs";
export const INTEGRATION_EDITOR_VIM = "VIM";

export interface RepoEventProps {
	repo: string;
	rev: string | null;
}
export interface FileEventProps extends RepoEventProps {
	path: string;
}
export interface SymbolEventProps extends FileEventProps {
	symbol: string;
	startLineNumber: number;
	endLineNumber: number;
	startColumn?: number;
	endColumn?: number;
}

// Interface to ensure no properties with common names are passed into logEvent with types that our analytics DB isn't expecting,
// such as non-string repo, rev, or path properties
interface ProtectedEventPropTypes {
	repo?: string;
	rev?: string | null;
	path?: string;
	// Any other unspecified property types are permitted
	[key: string]: any;
}

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

	logEvent(props?: ProtectedEventPropTypes): void {
		EventLogger.logEventWithComponents(this.category, this.action, this.label, props);
	}
}
export class NonInteractionLoggableEvent extends LoggableEvent {
	logEvent(props?: ProtectedEventPropTypes): void {
		EventLogger.logNonInteractionEventWithComponents(this.category, this.action, this.label, props);
	}
}
export function LogUnknownEvent(eventLabel: string): void {
	EventLogger.logEventWithComponents(EventCategories.Unknown, EventActions.Fetch, eventLabel);
}
// TODO (dadlerj): confirm props doesn't contain any type-unsafe entries
export function LogUnknownRedirectEvent(eventLabel: string, props?: any): void {
	EventLogger.logEventWithComponents(EventCategories.Unknown, EventActions.Fetch, eventLabel, props);
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
	Signup_Completed: new LoggableEvent("SignupCompleted", EventCategories.Auth, EventActions.Signup),
	Logout_Clicked: new LoggableEvent("LogoutClicked", EventCategories.Auth, EventActions.Logout),

	// Modals
	JoinModal_Initiated: new LoggableEvent("ShowSignUpModal", EventCategories.Auth, EventActions.Toggle),
	LoginModal_Initiated: new LoggableEvent("ShowLoginModal", EventCategories.Auth, EventActions.Toggle),
	ToolsModal_Initiated: new LoggableEvent("ClickToolsandIntegrations", EventCategories.Auth, EventActions.Toggle),
	BetaModal_Initiated: new LoggableEvent("ClickJoinBeta", EventCategories.Auth, EventActions.Toggle),
	AfterSignupModal_Initiated: new LoggableEvent("ShowAfterSignupModal", EventCategories.GTM, EventActions.Toggle),
	ChangePlanModal_Initiated: new LoggableEvent("ShowChangePlanModal", EventCategories.Billing, EventActions.Toggle),
	CancelSubscriptionModal_Initiated: new LoggableEvent("ShowCancelSubscriptionModal", EventCategories.Billing, EventActions.Toggle),
	CompleteSubscriptionModal_Initiated: new LoggableEvent("ShowCompleteSubscriptionModal", EventCategories.Billing, EventActions.Toggle),
	JoinModal_Dismissed: new LoggableEvent("DismissJoinModal", EventCategories.Auth, EventActions.Close),
	LoginModal_Dismissed: new LoggableEvent("DismissLoginModal", EventCategories.Auth, EventActions.Close),
	ToolsModal_Dismissed: new LoggableEvent("DismissToolsandIntegrationsModal", EventCategories.Auth, EventActions.Close),
	BetaModal_Dismissed: new LoggableEvent("DismissBetaModal", EventCategories.Auth, EventActions.Close),
	AfterSignupModal_Dismissed: new LoggableEvent("DismissAfterSignupModal", EventCategories.GTM, EventActions.Close),
	ChangePlanModal_Dismissed: new LoggableEvent("DismissChangePlanModal", EventCategories.Billing, EventActions.Close),
	CancelSubscriptionModal_Dismissed: new LoggableEvent("DismissCancelSubscriptionModal", EventCategories.Billing, EventActions.Close),
	CompleteSubscriptionModal_Dismissed: new LoggableEvent("DismissCompleteSubscriptionModal", EventCategories.Billing, EventActions.Close),

	// Toast events
	ToastChromeCTA_Clicked: new LoggableEvent("ChromeToastCTAClicked", EventCategories.Toast, EventActions.Click),
	ToastChrome_Closed: new LoggableEvent("ChromeToastCloseClicked", EventCategories.Toast, EventActions.Close),

	// Dashboard Events
	DashboardRepositoryTab_Clicked: new LoggableEvent("DashboardRepositoryTab_Clicked", EventCategories.Nav, EventActions.Click),

	// Repo events
	Repository_Clicked: new LoggableEvent("RepoClicked", EventCategories.Repository, EventActions.Click),
	RepositoryAuthedLanguagesGitHub_Fetched: new LoggableEvent("AuthedLanguagesGitHubFetched", EventCategories.Repository, EventActions.Fetch),
	RepositoryAuthedReposGitHub_Fetched: new LoggableEvent("AuthedReposGitHubFetched", EventCategories.Repository, EventActions.Fetch),

	// Repo page clicked
	ReposPageDocsButton_Clicked: new LoggableEvent("ReposPageDocsButtonClicked", EventCategories.Dashboard, EventActions.Click),
	ReposPageVideoButton_Clicked: new LoggableEvent("ReposPageVideoButtonClicked", EventCategories.Dashboard, EventActions.Click),
	ReposPageContactButton_Clicked: new LoggableEvent("ReposPageContactButtonClicked", EventCategories.Dashboard, EventActions.Click),

	// Signup
	SignupStage_Initiated: new LoggableEvent("SignupStageInitiated", EventCategories.Onboarding, EventActions.Toggle),
	SignupPlan_Selected: new LoggableEvent("SignupPlanSelected", EventCategories.Onboarding, EventActions.Click),
	SignupUserDetails_Completed: new LoggableEvent("SignupUserDetailsCompleted", EventCategories.Onboarding, EventActions.Success),
	SignupOrg_Selected: new LoggableEvent("SignupOrgSelected", EventCategories.Onboarding, EventActions.Click),
	SignupInfo_Completed: new LoggableEvent("SignupInfoCompleted", EventCategories.Onboarding, EventActions.Success),
	SignupEnterpriseForm_Completed: new LoggableEvent("SignupEnterpriseFormCompleted", EventCategories.Onboarding, EventActions.Success),
	AfterSignupModal_Completed: new LoggableEvent("AfterSignupModalCompleted", EventCategories.GTM, EventActions.Success),

	// Onboarding
	OnboardingRefsCoachCTA_Clicked: new LoggableEvent("ReferencesCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingJ2DCoachCTA_Clicked: new LoggableEvent("JumpToDefCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingSearchCoachCTA_Clicked: new LoggableEvent("SearchCoachmarkCTAClicked", EventCategories.Onboarding, EventActions.Click),
	OnboardingTour_Dismissed: new LoggableEvent("DismissTourCTAClicked", EventCategories.Onboarding, EventActions.Close),

	ChromeExtension_Installed: new LoggableEvent("ChromeExtensionInstalled", EventCategories.Onboarding, EventActions.Success),
	ChromeExtensionInstall_Failed: new LoggableEvent("ChromeExtensionInstallFailed", EventCategories.Onboarding, EventActions.Error),
	ChromeExtensionCTA_Clicked: new LoggableEvent("ChromeExtensionCTAClicked", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionInstall_Started: new LoggableEvent("ChromeExtensionInstallStarted", EventCategories.Onboarding, EventActions.Click),
	ChromeExtensionStore_Redirected: new LoggableEvent("ChromeExtensionStoreRedirect", EventCategories.Onboarding, EventActions.Click),

	// ReEngagement
	BetaSubscription_Completed: new LoggableEvent("BetaSubscriptionCompleted", EventCategories.ReEngagement, EventActions.Success),

	// Billing
	CancelSubscription_Clicked: new LoggableEvent("CancelSubscriptionClicked", EventCategories.Billing, EventActions.Click),
	ChangeSubscriptionRequest_Completed: new LoggableEvent("ChangeSubscriptionRequestCompleted", EventCategories.Billing, EventActions.Success),

	// Code view
	// Code view: Symbol events
	CodeContextMenu_Initiated: new LoggableEvent("CodeContextMenuClicked", EventCategories.CodeView, EventActions.Click),
	CodeExternalReferences_Viewed: new LoggableEvent("ClickedViewExternalReferences", EventCategories.CodeView, EventActions.Click),
	CodeToken_Hovered: new LoggableEvent("Hovering", EventCategories.CodeView, EventActions.Hover),
	CodeToken_Clicked: new LoggableEvent("BlobTokenClicked", EventCategories.CodeView, EventActions.Click),
	// Code view: Header events
	OpenInCodeHost_Clicked: new LoggableEvent("OpenInCodeHostClicked", EventCategories.CodeView, EventActions.Click),
	OpenInEditor_Clicked: new LoggableEvent("OpenInEditorClicked", EventCategories.CodeView, EventActions.Click),
	AuthorsToggle: new LoggableEvent("AuthorshipToggled", EventCategories.CodeView, EventActions.Toggle),
	// Code view: FileTree events
	FileTree_Navigated: new LoggableEvent("FileTreeActivated", EventCategories.CodeView, EventActions.Click),
	// Code view: CodeLens events
	CodeLensCommit_Clicked: new LoggableEvent("ClickedCodeLensCommit", EventCategories.CodeView, EventActions.Click),
	CodeLensCommitRedirect_Clicked: new LoggableEvent("ClickedCodeLensCommitRedirect", EventCategories.CodeView, EventActions.Click),
	// Code view: InfoPanel events
	InfoPanel_Initiated: new LoggableEvent("InfoPanelInitiated", EventCategories.CodeView, EventActions.Toggle),
	InfoPanel_Dismissed: new LoggableEvent("InfoPanelDismissed", EventCategories.CodeView, EventActions.Close),
	InfoPanelJumpToDef_Clicked: new LoggableEvent("InfoPanelJumpToDefClicked", EventCategories.CodeView, EventActions.Click),
	InfoPanelLocalRef_Toggled: new LoggableEvent("InfoPanelLocalRefDisplayed", EventCategories.CodeView, EventActions.Toggle),
	InfoPanelExternalRef_Toggled: new LoggableEvent("InfoPanelExternalRefDisplayed", EventCategories.CodeView, EventActions.Toggle),
	InfoPanelRefPreview_Closed: new LoggableEvent("InfoPanelRefPreviewClosed", EventCategories.CodeView, EventActions.Close),
	InfoPanelRefPreviewTitle_Clicked: new LoggableEvent("InfoPanelRefPreviewTitleClicked", EventCategories.CodeView, EventActions.Click),
	InfoPanelComment_Toggled: new LoggableEvent("InfoPanelCommentToggled", EventCategories.CodeView, EventActions.Toggle),
	// Code view: CommitInfoBar events
	CommitInfoItem_Selected: new LoggableEvent("CommitInfoItemSelected", EventCategories.CodeView, EventActions.Click),
	CommitInfo_Initiated: new LoggableEvent("CommitInfoInitiated", EventCategories.CodeView, EventActions.Toggle),
	CommitInfo_Dismissed: new LoggableEvent("CommitInfoDismissed", EventCategories.CodeView, EventActions.Close),
	// Quick open/search
	QuickopenItem_Selected: new LoggableEvent("QuickOpenItemSelected", EventCategories.QuickOpen, EventActions.Click),
	Quickopen_Initiated: new LoggableEvent("QuickOpenInitiated", EventCategories.QuickOpen, EventActions.Toggle),
	Quickopen_Dismissed: new LoggableEvent("QuickOpenDismissed", EventCategories.QuickOpen, EventActions.Close),

	ShortcutMenu_Initiated: new LoggableEvent("ShorcutMenuInitiated", EventCategories.ShortcutMenu, EventActions.Toggle),
	ShortcutMenu_Dismissed: new LoggableEvent("ShorcutMenuDismissed", EventCategories.ShortcutMenu, EventActions.Close),

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

	// Redirect/external events
	RepoBadge_Redirected: new LoggableEvent("RepoBadgeRedirected", EventCategories.External, EventActions.Redirect),

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
		"afterSignup": Events.AfterSignupModal_Dismissed,
		"cancelSubscriptionModal": Events.CancelSubscriptionModal_Dismissed,
		"trialCompletionModal": Events.CompleteSubscriptionModal_Dismissed,
		"planChanger": Events.ChangePlanModal_Dismissed,
		"keyboardShortcuts": Events.ShortcutMenu_Dismissed,
	};
	return (modalName && modalName in dismissModalsMap) ? dismissModalsMap[modalName] : null;
}
