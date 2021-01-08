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

        // debugging
        // TODO(tj): remove after full refactor complete
        this.panelViews
            .getPanelViews(ContributableViewContainer.Panel)
            .subscribe(panels => console.log('registered panel views', panels))
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

    // TODO general services
    public readonly contribution = new ContributionRegistry(
        this.viewer,
        this.model,
        this.platformContext.settings,
        this.context.data
    )
    public readonly view = createViewService()
    public readonly extensions: IExtensionsService

    // Feature provider services
    public readonly textDocumentReferences = new TextDocumentLocationProviderRegistry<ReferenceParameters>()
    public readonly textDocumentLocations = new TextDocumentLocationProviderIDRegistry()
    public readonly textDocumentDecoration = new TextDocumentDecorationProviderRegistry()

    public readonly linkPreviews = new LinkPreviewProviderRegistry()
    public readonly panelViews = new PanelViewProviderRegistry()
    public readonly completionItems = new CompletionItemProviderRegistry()
}
