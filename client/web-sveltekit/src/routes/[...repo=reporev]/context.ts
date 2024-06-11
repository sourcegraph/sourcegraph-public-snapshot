/**
 * This context allows repository pages to propagte relevant information to other components.
 * Right now it is used to generate context-related suggestions in the repository search input.
 */
export interface RepositoryPageContext {
    revision?: string
    filePath?: string
    directoryPath?: string
    fileLanguage?: string
}
