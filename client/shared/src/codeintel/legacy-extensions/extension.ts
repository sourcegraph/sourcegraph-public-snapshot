import * as sourcegraph from './api'

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

/**
 * Activate the extension by registering definition, reference, and hover providers
 * with LSIF and search-based providers.
 *
 * @param context  The extension context.
 */
export const activate = async (context: sourcegraph.ExtensionContext = DUMMY_CTX): Promise<void> => {
    // const languageSpec = languageSpecs.find(spec => spec.languageID === languageID)
    // if (languageSpec === undefined) {
    //     throw new Error(`Unknown language ${languageID}`)
    // }
    // const selector: sourcegraph.DocumentSelector = [
    //     { language: languageSpec.languageID },
    //     ...(languageSpec.verbatimFilenames || []).flatMap(filename => [{ pattern: filename }]),
    //     ...languageSpec.fileExts.flatMap(extension => [{ pattern: `*.${extension}` }]),
    // ]
    // const hasImplementationsFieldConst = true
    // const providers = createProviders(languageSpec, hasImplementationsFieldConst, new RedactingLogger(console))
    // context.subscriptions.add(sourcegraph.languages.registerDefinitionProvider(selector, providers.definition))
    // context.subscriptions.add(sourcegraph.languages.registerHoverProvider(selector, providers.hover))
    // context.subscriptions.add(
    //     sourcegraph.languages.registerDocumentHighlightProvider(selector, providers.documentHighlights)
    // )
    // context.subscriptions.add(sourcegraph.languages.registerReferenceProvider(selector, providers.references))
    // if (hasImplementationsFieldConst) {
    //     if (languageSpec.textDocumentImplemenationSupport) {
    //         // Show the "Find implementations" button in the hover as specified in package.json (look for
    //         // "findImplementations").
    //         sourcegraph.internal.updateContext({
    //             ['implementations_LANGID']: true,
    //         })
    //         // Create an Implementations panel and register a locations provider.
    //         createImplementationPanel(context, selector, providers)
    //     }
    // }
}
