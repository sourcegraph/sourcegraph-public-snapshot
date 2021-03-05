import { PlatformContext } from '../../platform/context'
import { ContributableViewContainer, ReferenceParameters } from '../protocol'
import { createContextService } from './context/contextService'
import { CompletionItemProviderRegistry } from './services/completion'
import { ContributionRegistry } from './services/contribution'
import { TextDocumentDecorationProviderRegistry } from './services/decoration'
import { createViewerService } from './services/viewerService'
import { IExtensionsService, ExtensionsService } from './services/extensionsService'
import { LinkPreviewProviderRegistry } from './services/linkPreview'
import { TextDocumentLocationProviderIDRegistry, TextDocumentLocationProviderRegistry } from './services/location'
import { createModelService } from './services/modelService'
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
    ) {
        if (platformContext.createExtensionsService) {
            this.extensions = platformContext.createExtensionsService(this.model)
        } else {
            this.extensions = new ExtensionsService(platformContext, this.model)
        }
    }

    // TEMP MIGRATION CHECKLIST
    // moved to mainthread
    // public readonly commands = new CommandRegistry()
    // moved to exthost
    public readonly context = createContextService(this.platformContext)
    // moved to ext host
    public readonly model = createModelService()
    // moved to ext host
    public readonly viewer = createViewerService(this.model)
    // notifications have moved to both main thread and ext host (depending on whether they need user input)
    // moved to ext host
    public readonly workspace = createWorkspaceService()
    public readonly extensions: IExtensionsService
    public readonly contribution = new ContributionRegistry(
        this.viewer,
        this.model,
        this.platformContext.settings,
        this.context.data
    )
    public readonly textDocumentDecoration = new TextDocumentDecorationProviderRegistry()
    public readonly textDocumentReferences = new TextDocumentLocationProviderRegistry<ReferenceParameters>()
    public readonly textDocumentLocations = new TextDocumentLocationProviderIDRegistry()
    public readonly panelViews = new PanelViewProviderRegistry()

    // Feature provider services

    // TODO(tj): refactor this in a separate PR
    public readonly view = createViewService()

    public readonly linkPreviews = new LinkPreviewProviderRegistry() // TODO(tj): remove and deprecate
    public readonly completionItems = new CompletionItemProviderRegistry()
}
