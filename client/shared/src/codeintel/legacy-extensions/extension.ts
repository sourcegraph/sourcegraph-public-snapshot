import { once } from 'lodash'
import { from } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'
import * as sourcegraph from '@sourcegraph/extension-api-types'

import { languageID } from './language'
import { languageSpecs } from './language-specs/languages'
import { RedactingLogger } from './logging'
import { createProviders, SourcegraphProviders } from './providers'
import { API } from './util/api'

/**
 * A dummy context that is used for versions of Sourcegraph to 3.0.
 */
const DUMMY_CTX = {
    subscriptions: {
        add: (): void => {
            /* no-op */
        },
    },
}

const hasImplementationsField = once(() => new API().hasImplementationsField())

/**
 * Create the panel for implementations.
 *
 * Makes sure to only create the panel once per session.
 */
const createImplementationPanel = (
    context: sourcegraph.ExtensionContext = DUMMY_CTX,
    selector: sourcegraph.DocumentSelector,
    providers: SourcegraphProviders
): void => {
    const implementationsPanelID = 'implementations_LANGID'
    const implementationsPanel = sourcegraph.app.createPanelView(implementationsPanelID)

    implementationsPanel.title = 'Implementations'
    implementationsPanel.component = { locationProvider: implementationsPanelID }
    implementationsPanel.priority = 160
    implementationsPanel.selector = selector

    const maxPanelResults = sourcegraph.configuration
        .get<{ 'codeIntelligence.maxPanelResults'?: number }>()
        .get('codeIntelligence.maxPanelResults')
    if (maxPanelResults) {
        implementationsPanel.component.maxLocationResults = maxPanelResults
    }

    context.subscriptions.add(implementationsPanel)
    context.subscriptions.add(
        sourcegraph.languages.registerLocationProvider(implementationsPanelID, selector, providers.implementations)
    )
}

/**
 * Activate the extension by registering definition, reference, and hover providers
 * with LSIF and search-based providers.
 *
 * @param context  The extension context.
 */
export const activate = async (context: sourcegraph.ExtensionContext = DUMMY_CTX): Promise<void> => {
    const languageSpec = languageSpecs.find(spec => spec.languageID === languageID)
    if (languageSpec === undefined) {
        throw new Error(`Unknown language ${languageID}`)
    }

    const selector: sourcegraph.DocumentSelector = [
        { language: languageSpec.languageID },
        ...(languageSpec.verbatimFilenames || []).flatMap(filename => [{ pattern: filename }]),
        ...languageSpec.fileExts.flatMap(extension => [{ pattern: `*.${extension}` }]),
    ]

    const hasImplementationsFieldConst = await hasImplementationsField()

    const providers = createProviders(languageSpec, hasImplementationsFieldConst, new RedactingLogger(console))
    context.subscriptions.add(sourcegraph.languages.registerDefinitionProvider(selector, providers.definition))
    context.subscriptions.add(sourcegraph.languages.registerHoverProvider(selector, providers.hover))

    // Do not try to register this provider on pre-3.18 instances as
    // it didn't exist.
    if (sourcegraph.languages.registerDocumentHighlightProvider) {
        context.subscriptions.add(
            sourcegraph.languages.registerDocumentHighlightProvider(selector, providers.documentHighlights)
        )
    }

    // Re-register the references provider whenever the value of the
    // mixPreciseAndSearchBasedReferences setting changes.

    let unsubscribeReferencesProvider: sourcegraph.Unsubscribable
    const registerReferencesProvider = (): void => {
        unsubscribeReferencesProvider?.unsubscribe()
        unsubscribeReferencesProvider = sourcegraph.languages.registerReferenceProvider(selector, providers.references)
        context.subscriptions.add(unsubscribeReferencesProvider)
    }

    context.subscriptions.add(
        from(sourcegraph.configuration)
            .pipe(
                startWith(false),
                map(() => sourcegraph.configuration.get().get('codeIntel.mixPreciseAndSearchBasedReferences') ?? false),
                distinctUntilChanged(),
                map(registerReferencesProvider)
            )
            .subscribe()
    )

    if (hasImplementationsFieldConst) {
        if (languageSpec.textDocumentImplemenationSupport) {
            // Show the "Find implementations" button in the hover as specified in package.json (look for
            // "findImplementations").
            sourcegraph.internal.updateContext({
                ['implementations_LANGID']: true,
            })

            // Create an Implementations panel and register a locations provider.
            createImplementationPanel(context, selector, providers)
        }
    }
}
