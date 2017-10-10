import { EventActions, EventCategories } from './analyticsConstants'
import { pageViewQueryParameters } from './analyticsUtils'
import { eventLogger } from './eventLogger'

export class LoggableViewEvent {
    constructor(private title: string) { }
    public log(props?: any): void {
        eventLogger.logViewEvent(this.title, { ...props, ...pageViewQueryParameters(window.location.href) })
    }
}

/**
 * Loggable pageview events to be used throughout the application
 *
 * Note: all newly added events should follow the "View$Page" naming scheme
 */
export const viewEvents = {
    Home: new LoggableViewEvent('ViewHome'),
    SearchResults: new LoggableViewEvent('ViewSearchResults'),
    Blob: new LoggableViewEvent('ViewBlob'),
    UserProfile: new LoggableViewEvent('ViewUserProfile'),
    SignIn: new LoggableViewEvent('ViewSignIn'),
    EditorAuth: new LoggableViewEvent('ViewEditorAuth'),
    AddNewOrg: new LoggableViewEvent('ViewAddNewOrg'),
    OrgProfile: new LoggableViewEvent('ViewOrgProfile')
}

export class LoggableEvent {
    constructor(private eventLabel: string, private eventCategory: string, private eventAction: string) { }
    public log(props?: any): void {
        eventLogger.logEvent(this.eventCategory, this.eventAction, this.eventLabel, props)
    }
}

/**
 * Loggable events to be used throughout the application
 *
 * Note: all newly added events should follow the "$Noun$Verb" naming scheme
 */
export const events = {
    // Auth
    InitiateSignIn: new LoggableEvent('InitiateSignIn', EventCategories.Auth, EventActions.Click),
    InitiateSignUp: new LoggableEvent('InitiateSignUp', EventCategories.Auth, EventActions.Click),
    SignupCompleted: new LoggableEvent('SignupCompleted', EventCategories.Auth, EventActions.SignUp),
    SignOutClicked: new LoggableEvent('SignOutClicked', EventCategories.Auth, EventActions.Click),
    CompletedAuth0SignIn: new LoggableEvent('CompletedAuth0SignIn', EventCategories.Auth, EventActions.SignIn),
    EditorAuthIdCopied: new LoggableEvent('EditorAuthIdCopied', EventCategories.Auth, EventActions.Click),

    // Settings events
    CreateNewOrgClicked: new LoggableEvent('CreateNewOrgClicked', EventCategories.Settings, EventActions.Click),
    NewOrgFailed: new LoggableEvent('NewOrgFailed', EventCategories.Settings, EventActions.Error),
    NewOrgCreated: new LoggableEvent('NewOrgCreated', EventCategories.Settings, EventActions.Success),

    NewUserFailed: new LoggableEvent('NewUserFailed', EventCategories.Settings, EventActions.Error),
    NewUserCreated: new LoggableEvent('NewUserCreated', EventCategories.Settings, EventActions.Success),
    UpdateUserClicked: new LoggableEvent('UpdateUserClicked', EventCategories.Settings, EventActions.Success),
    UpdateUserFailed: new LoggableEvent('UpdateUserFailed', EventCategories.Settings, EventActions.Error),

    InviteOrgMemberClicked: new LoggableEvent('InviteOrgMemberClicked', EventCategories.Settings, EventActions.Click),
    InviteOrgMemberFailed: new LoggableEvent('InviteOrgMemberFailed', EventCategories.Settings, EventActions.Error),
    OrgMemberInvited: new LoggableEvent('OrgMemberInvited', EventCategories.Settings, EventActions.Success),
    AcceptInviteFailed: new LoggableEvent('AcceptInviteFailed', EventCategories.Settings, EventActions.Error),
    InviteAccepted: new LoggableEvent('InviteAccepted', EventCategories.Settings, EventActions.Success),

    RemoveOrgMemberClicked: new LoggableEvent('RemoveOrgMemberClicked', EventCategories.Settings, EventActions.Click),
    RemoveOrgMemberFailed: new LoggableEvent('RemoveOrgMemberFailed', EventCategories.Settings, EventActions.Error),
    OrgMemberRemoved: new LoggableEvent('OrgMemberRemoved', EventCategories.Settings, EventActions.Success),

    // Nav bar events
    ShareButtonClicked: new LoggableEvent('ShareButtonClicked', EventCategories.Sharing, EventActions.Click),
    OpenInCodeHostClicked: new LoggableEvent('OpenInCodeHostClicked', EventCategories.External, EventActions.Click),
    OpenInNativeAppClicked: new LoggableEvent('OpenInNativeAppClicked', EventCategories.External, EventActions.Click),

    // Blob view events
    SymbolHovered: new LoggableEvent('SymbolHovered', EventCategories.Editor, EventActions.Hover),
    TooltipDocked: new LoggableEvent('TooltipDocked', EventCategories.Editor, EventActions.Click),
    TextSelected: new LoggableEvent('TextSelected', EventCategories.Editor, EventActions.Select),
    GoToDefClicked: new LoggableEvent('GoToDefClicked', EventCategories.Editor, EventActions.Click),
    FindRefsClicked: new LoggableEvent('FindRefsClicked', EventCategories.Editor, EventActions.Click),
    SearchClicked: new LoggableEvent('SearchClicked', EventCategories.Editor, EventActions.Click),

    // Refs panel events
    ShowAllRefsButtonClicked: new LoggableEvent('ShowAllRefsButtonClicked', EventCategories.Editor, EventActions.Click),
    ShowLocalRefsButtonClicked: new LoggableEvent('ShowLocalRefsButtonClicked', EventCategories.Editor, EventActions.Click),
    ShowExternalRefsButtonClicked: new LoggableEvent('ShowExternalRefsButtonClicked', EventCategories.Editor, EventActions.Click),
    GoToLocalRefClicked: new LoggableEvent('GoToLocalRefClicked', EventCategories.Editor, EventActions.Click),
    GoToExternalRefClicked: new LoggableEvent('GoToExternalRefClicked', EventCategories.Editor, EventActions.Click),

    // Search events
    SearchSubmitted: new LoggableEvent('SearchSubmitted', EventCategories.Search, EventActions.Submit),

    // External events
    RepoBadgeRedirected: new LoggableEvent('RepoBadgeRedirected', EventCategories.External, EventActions.Redirect)
}
