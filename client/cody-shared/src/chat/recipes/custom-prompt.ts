import type { URI } from 'vscode-uri'

import type { CodebaseContext } from '../../codebase-context'
import type { ContextMessage } from '../../codebase-context/messages'
import type { ActiveTextEditorSelection, Editor } from '../../editor'
import { MAX_HUMAN_INPUT_TOKENS, NUM_CODE_RESULTS, NUM_TEXT_RESULTS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import type { CodyPromptContext } from '../prompts'
import {
    extractTestType,
    getHumanLLMText,
    isOnlySelectionRequired,
    newInteraction,
    newInteractionWithError,
} from '../prompts/utils'
import {
    getCurrentDirContext,
    getCurrentFileContextFromEditorSelection,
    getCurrentFileImportsContext,
    getDirectoryFileListContext,
    getEditorDirContext,
    getEditorOpenTabsContext,
    getFilePathContext,
    getHumanDisplayTextWithFileName,
    getPackageJsonContext,
    getTerminalOutputContext,
} from '../prompts/vscode-context'
import type { Interaction } from '../transcript/interaction'

import { getFileExtension, numResults } from './helpers'
import type { Recipe, RecipeContext, RecipeID } from './recipe'

/**
 * ======================================================
 * Recipe for running custom prompts from the cody.json files
 * Works with VS Code only
 ====================================================== *
 */
export class CustomPrompt implements Recipe {
    public id: RecipeID = 'custom-prompt'
    public title = 'Custom Prompt'

    /**
     * Retrieves an Interaction object based on the humanChatInput and RecipeContext provided.
     * The Interaction object contains messages from both the human and the assistant, as well as context information.
     */
    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const workspaceRootUri = context.editor.getWorkspaceRootUri()
        // Check if context is required
        const contextConfigString = (await context.editor.controllers?.command?.get('context')) || ''
        const contextConfig = JSON.parse(contextConfigString) as CodyPromptContext

        // If selection is required, ensure not to accept visible content as selection
        const selection = contextConfig?.selection
            ? context.editor.getActiveTextEditorSelection()
            : context.editor.getActiveTextEditorSelectionOrVisibleContent()

        const command = context.editor.controllers?.command?.getCurrentCommand()

        // Get prompt text from the editor command or from the human input
        const promptText = humanChatInput.trim() || command?.prompt
        if (!promptText) {
            const errorMessage = 'Please enter a valid prompt for the custom command.'
            return newInteractionWithError(errorMessage, promptText || '')
        }

        const commandName = command?.slashCommand || command?.description || promptText
        if (contextConfig?.selection && !selection?.selectedText) {
            const errorMessage = `__${commandName}__ requires highlighted code. Please select some code in your editor and try again.`
            return newInteractionWithError(errorMessage, commandName)
        }

        // Add selection file name as display when available
        const displayText = getHumanDisplayTextWithFileName(commandName, selection, workspaceRootUri)
        const text = getHumanLLMText(promptText, selection?.fileName)

        // Attach code selection to prompt text if only selection is needed as context
        if (selection && isOnlySelectionRequired(contextConfig)) {
            const contextMessages = Promise.resolve(getCurrentFileContextFromEditorSelection(selection))
            return newInteraction({ text, displayText, contextMessages })
        }

        // Get output from the command if any
        const commandOutput = await context.editor.controllers?.command?.get('output')

        const truncatedText = truncateText(text, MAX_HUMAN_INPUT_TOKENS)
        const contextMessages = this.getContextMessages(
            truncatedText,
            context.editor,
            context.codebaseContext,
            contextConfig,
            selection,
            commandOutput
        )

        return newInteraction({ text: truncatedText, displayText, contextMessages })
    }

    private async getContextMessages(
        text: string,
        editor: Editor,
        codebaseContext: CodebaseContext,
        promptContext: CodyPromptContext,
        selection?: ActiveTextEditorSelection | null,
        commandOutput?: string | null
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []
        const workspaceRootUri = editor.getWorkspaceRootUri()
        const isUnitTestRequest = extractTestType(text) === 'unit'

        if (promptContext.none) {
            return []
        }

        if (promptContext.codebase) {
            const codebaseMessages = await codebaseContext.getContextMessages(text, numResults)
            contextMessages.push(...codebaseMessages)
        }
        if (promptContext.openTabs) {
            const openTabsMessages = await getEditorOpenTabsContext()
            contextMessages.push(...openTabsMessages)
        }
        if (promptContext.currentDir) {
            const currentDirMessages = await getCurrentDirContext(isUnitTestRequest)
            contextMessages.push(...currentDirMessages)
        }
        if (!promptContext.directoryPath) {
            if (promptContext.directoryPath) {
                const dirMessages = await getEditorDirContext(promptContext.directoryPath, selection?.fileName)
                contextMessages.push(...dirMessages)
            }
        }
        if (!promptContext.filePath) {
            if (promptContext.filePath) {
                const fileMessages = await getFilePathContext(promptContext.filePath)
                contextMessages.push(...fileMessages)
            }
        }

        // Context for unit tests requests
        if (isUnitTestRequest && contextMessages.length === 0) {
            if (selection?.fileName) {
                const importsContext = await this.getUnitTestContextMessages(selection, workspaceRootUri)
                contextMessages.push(...importsContext)
            }
        }

        if (promptContext.currentFile || promptContext.selection !== false) {
            if (selection) {
                const currentFileMessages = getCurrentFileContextFromEditorSelection(selection)
                contextMessages.push(...currentFileMessages)
            }
        }
        if (promptContext.command && commandOutput) {
            const outputMessages = getTerminalOutputContext(commandOutput)
            contextMessages.push(...outputMessages)
        }
        // Return sliced results
        const maxResults = Math.floor((NUM_CODE_RESULTS + NUM_TEXT_RESULTS) / 2) * 2
        return contextMessages.slice(-maxResults * 2)
    }

    private async getUnitTestContextMessages(
        selection: ActiveTextEditorSelection,
        workspaceRootUri?: URI | null
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []

        if (workspaceRootUri) {
            const rootFileNames = await getDirectoryFileListContext(workspaceRootUri, true)
            contextMessages.push(...rootFileNames)
        }
        // Add package.json content only if files matches the ts/js extension regex
        if (selection?.fileName && getFileExtension(selection?.fileName).match(/ts|js/)) {
            const packageJson = await getPackageJsonContext(selection?.fileName)
            contextMessages.push(...packageJson)
        }
        // Try adding import statements from current file as context
        if (selection?.fileName) {
            const importsContext = await getCurrentFileImportsContext()
            contextMessages.push(...importsContext)
        }

        return contextMessages
    }
}
