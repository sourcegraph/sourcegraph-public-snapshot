import * as vscode from 'vscode'

import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { ActiveTextEditorSelection } from '../../editor'
import {
    MAX_CURRENT_FILE_TOKENS,
    MAX_HUMAN_INPUT_TOKENS,
    MAX_RECIPE_INPUT_TOKENS,
    MAX_RECIPE_SURROUNDING_TOKENS,
} from '../../prompt/constants'
import { populateCurrentEditorContextTemplate } from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { BufferedBotResponseSubscriber } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

import { ChatQuestion } from './chat-question'
import { commandRegex, contentSanitizer } from './helpers'
import { Recipe, RecipeContext, RecipeID } from './recipe'

/** ======================================================
 * Recipe for Generating a New File
====================================================== **/
export class InlineTouch implements Recipe {
    public id: RecipeID = 'inline-touch'
    private workspacePath = vscode.workspace.workspaceFolders?.[0].uri

    constructor(private debug: (filterLabel: string, text: string, ...args: unknown[]) => void) {}

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelection() || context.editor.controllers?.inline.selection
        if (!selection || !this.workspacePath) {
            await context.editor.controllers?.inline.error()
            await context.editor.showWarningMessage('Failed to start Inline Chat: empty selection.')
            return null
        }
        const humanInput = humanChatInput.trim() || (await this.getInstructionFromInput()).trim()
        if (!humanInput) {
            await context.editor.controllers?.inline.error()
            await context.editor.showWarningMessage('Failed to start Inline Chat: empty input.')
            return null
        }
        // Get the current directory of the file that the user is currently working on
        // Create file path from selection.fileName and workspacePath
        const currentFilePath = `${this.workspacePath.fsPath}/${selection.fileName}`
        const currentDir = currentFilePath.replace(/\/[^/]+$/, '')
        this.debug('InlineTouch:currentDir', 'currentDir', currentDir)

        // Create new file name based on the user's input
        const newFileName = commandRegex.noTest.test(humanInput)
            ? currentFilePath.replace(/(\.[^./]+)$/, '.cody$1')
            : currentFilePath.replace(/(\.[^./]+)$/, '.test$1')
        const newFsPath = newFileName || (await this.getNewFileNameFromInput(selection.fileName, currentDir))
        if (!newFsPath || !currentDir) {
            return null
        }

        // create vscode uri for the new file from the newFilePath which includes the workspacePath
        const fileUri = vscode.Uri.file(newFsPath)
        const workspaceEditor = new vscode.WorkspaceEdit()
        // Create file if it doesn't exist
        workspaceEditor.createFile(fileUri, { ignoreIfExists: true })
        await vscode.workspace.applyEdit(workspaceEditor)
        this.debug('InlineTouch:workspaceEditor', 'createFile', fileUri)

        const truncatedText = truncateText(humanInput, MAX_HUMAN_INPUT_TOKENS)
        const MAX_RECIPE_CONTENT_TOKENS = MAX_RECIPE_INPUT_TOKENS + MAX_RECIPE_SURROUNDING_TOKENS * 2
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_CONTENT_TOKENS)

        // Reconstruct Cody's prompt using user's context
        // Replace placeholders in reverse order to avoid collisions if a placeholder occurs in the input
        const prompt = InlineTouch.newFilePrompt
        const promptText = prompt
            .replace('{newFileName}', newFsPath)
            .replace('{humanInput}', truncatedText)
            .replace('{selectedText}', truncatedSelectedText)
            .replace('{fileName}', selection.fileName)

        // Text display in UI fpr human that includes the selected code
        const displayText = this.getHumanDisplayText(humanInput, selection.fileName)
        context.responseMultiplexer.sub(
            'selection',
            new BufferedBotResponseSubscriber(async content => {
                if (!content) {
                    await context.editor.controllers?.inline.error()
                    await context.editor.showWarningMessage(
                        'Cody did not suggest any code updates. Please try again with a different question.'
                    )
                    return
                }
                await this.addContentToNewFile(workspaceEditor, fileUri, content)
                this.debug('InlineTouch:responseMultiplexer', 'BufferedBotResponseSubscriber', content)
            })
        )

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText,
                },
                {
                    speaker: 'assistant',
                    prefix: 'Working on it! I will show you the new file when it is ready.\n\n',
                },
                this.getContextMessages(selection, currentDir),
                []
            )
        )
    }

    private async addContentToNewFile(
        workspaceEditor: vscode.WorkspaceEdit,
        filePath: vscode.Uri,
        content: string
    ): Promise<void> {
        const textDocument = await vscode.workspace.openTextDocument(filePath)
        workspaceEditor.insert(filePath, new vscode.Position(textDocument.lineCount + 1, 0), contentSanitizer(content))
        await vscode.workspace.applyEdit(workspaceEditor)
        await textDocument.save()
        await vscode.window.showTextDocument(filePath)
    }

    /** ======================================================
     * Prompt Template for New File
     * ====================================================== */

    public static readonly newFilePrompt = `
    I am currently looking at this selected code from {fileName}:
    \`\`\`
    {selectedText}
    \`\`\`

    Help me with creating content for a new file based on the selected code.
    - {humanInput}

    ## Instruction
    - Follow my instructions to produce new code for the new file called {newFileName}.
    - Think carefully and use the shared context as reference before produce the new code
    - Make sure the new code works with the shared context and the selected code.
    - Use the same framework, language and style as the shared context that are also from current directory I am working on.
    - Put all new content inside <selection> tags.
    - I only want to see the new code enclosed with the <selection> tags only if you understand my instructions.
    - Do not enclose any part of your answer with <selection> tags if you are not sure about the answer.
    - Only provide me with the code inside <selection> and nothing else.
    - Do not enclose your answer with markdowns.
    ## Guidelines for the new file
    - Include all the import statements that are required for the new code to work.
    - If there are already content in the file with the same name, the new code will be appended to the file.
    - If my selected code is empty, it means I am working in an empty file.
    - Do not remove code that is being used by the the shared files.
    - Do not suggest code that are not related to any of the shared context.
    - Do not make up code, including function names, that could break the selected code.
    `

    // Prompt template for displaying the prompt to users in chat view
    public static readonly displayPrompt = `\n
    File: `

    // ======================================================== //
    //                      GET CONTEXT                         //
    // ======================================================== //

    private async getContextMessages(
        selection: ActiveTextEditorSelection,
        currentDir: string
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []
        // Add selected text and current file as context and create context messages from current directory
        const selectedContext = ChatQuestion.getEditorSelectionContext(selection)
        const currentDirContext = await InlineTouch.getEditorDirContext(currentDir)
        contextMessages.push(...selectedContext, ...currentDirContext)
        // Create context messages from open tabs
        if (contextMessages.length < 10) {
            contextMessages.push(...InlineTouch.getEditorOpenTabsContext(currentDir))
        }
        return contextMessages.slice(-10)
    }

    // Create Context from Current Directory of Active Document //
    public static async getEditorDirContext(currentDir: string): Promise<ContextMessage[]> {
        // get a list of files from the current directory path
        const currentDirUri = vscode.Uri.file(currentDir)
        // Get the list of files in the current directory then filter out directories and get the first 10 files
        const filesInDir = (await vscode.workspace.fs.readDirectory(currentDirUri))
            .filter(file => file[1] === 1)
            .slice(0, 10)
        const contextMessages: ContextMessage[] = []
        for (const file of filesInDir) {
            // Get the context from each file
            const fileName = vscode.Uri.joinPath(currentDirUri, file[0]).fsPath
            const fileUri = vscode.Uri.joinPath(currentDirUri, file[0])
            try {
                const fileContent = await vscode.workspace.openTextDocument(fileUri)
                const truncatedContent = truncateText(fileContent.getText(), MAX_CURRENT_FILE_TOKENS)
                const contextMessage = getContextMessageWithResponse(
                    populateCurrentEditorContextTemplate(truncatedContent, fileName),
                    { fileName }
                )
                contextMessages.push(...contextMessage)
            } catch (error) {
                console.error(error)
            }
        }

        return contextMessages
    }

    // Get context from current editor open tabs
    public static getEditorOpenTabsContext(currentDir: string): ContextMessage[] {
        const contextMessages: ContextMessage[] = []
        // Skip the current active tab (which is already included in selection context), files in currentDir, and non-file tabs
        const openTabs = vscode.window.visibleTextEditors
        for (const tab of openTabs) {
            if (
                tab === vscode.window.activeTextEditor ||
                tab.document.fileName.includes(currentDir) ||
                tab.document.uri.scheme !== 'file'
            ) {
                continue
            }
            const fileName = tab.document.fileName
            const truncatedContent = truncateText(tab.document.getText(), MAX_CURRENT_FILE_TOKENS)
            const contextMessage = getContextMessageWithResponse(
                populateCurrentEditorContextTemplate(truncatedContent, fileName),
                {
                    fileName,
                }
            )
            contextMessages.push(...contextMessage)
        }
        return contextMessages
    }

    // ======================================================== //
    //                          HELPERS                         //
    // ======================================================== //

    // Get display text for human
    private getHumanDisplayText(humanChatInput: string, fileName: string): string {
        return '**✨Touch✨** ' + humanChatInput + InlineTouch.displayPrompt + fileName
    }

    private async getInstructionFromInput(): Promise<string> {
        // Get the file name from the user using the input box, set default value to cody and validate the input
        const humanInput = await vscode.window.showInputBox({
            prompt: 'Enter your instructions for Cody to create a new file based on the selected code:',
            placeHolder: 'ex. create unit tests for the selected code',
            validateInput: (input: string) => {
                if (!input) {
                    return 'Please enter instructions.'
                }
                return null
            },
        })
        return humanInput || ''
    }

    private async getNewFileNameFromInput(fileName: string, currentDir: string): Promise<string> {
        // Get the file name from the user using the input box, set default value to cody and validate the input
        const newFileName = await vscode.window.showInputBox({
            prompt: 'Enter a new file name (with extension):',
            value: fileName,
            validateInput: (input: string) => {
                if (!input) {
                    return 'Please enter a file name.'
                }
                if (!input.includes('.')) {
                    return 'Please enter a file name with extension.'
                }
                return null
            },
        })
        // The newFilePath is the fsPath of the new file that the user is creating
        return `${currentDir}/${newFileName}`
    }
}
