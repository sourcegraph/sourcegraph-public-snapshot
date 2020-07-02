import { PlatformContext } from '../../platform/context'
import { ReferenceParams } from '../protocol'
import { createContextService } from './context/contextService'
import { CommandRegistry } from './services/command'
import { ContributionRegistry } from './services/contribution'
import { LinkPreviewProviderRegistry } from './services/linkPreview'
import { NotificationsService } from './services/notifications'
import { PanelViewProviderRegistry } from './services/panelViews'
import { createViewService } from './services/viewService'
import { createWorkspaceService } from './services/workspaceService'

/**
 * Services is a container for all services used by the client application.
 */
export class Services {
    constructor(
        private platformContext: Pick<
            PlatformContext,
            | 'settings'
            | 'updateSettings'
            | 'requestGraphQL'
            | 'getScriptURLForExtension'
            | 'clientApplication'
            | 'sideloadedExtensionURL'
            | 'createExtensionsService'
        >
    ) {}

    public readonly commands = new CommandRegistry()
    public readonly context = createContextService(this.platformContext)
    public readonly workspace = createWorkspaceService()
    public readonly notifications = new NotificationsService()
    public readonly contribution = new ContributionRegistry(
        this.viewer,
        this.model,
        this.platformContext.settings,
        this.context.data
    )
    public readonly linkPreviews = new LinkPreviewProviderRegistry()
    public readonly panelViews = new PanelViewProviderRegistry()
    public readonly view = createViewService()
}
