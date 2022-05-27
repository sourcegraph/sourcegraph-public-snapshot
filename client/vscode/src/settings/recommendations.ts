import * as vscode from 'vscode'

import { DISMISS_WORKSPACERECS_CTA_KEY, LocalStorageService } from './LocalStorageService'
/**
 * Ask if user wants to add Sourcegraph to their Workspace Recommendations list by displaying built-in popup
 * It will not show popup if the user already has Sourcegraph added to their recommendations list
 * or has dismissed the message previously
 */
export async function recommendSourcegraph(localStorageService: LocalStorageService): Promise<void> {
    // Check if user has clicked on dismiss previously
    // Reset stored value for testing:
    // await localStorageService.setValue(DISMISS_WORKSPACERECS_CTA_KEY, '')
    const isDismissed = localStorageService.getValue(DISMISS_WORKSPACERECS_CTA_KEY)
    if (isDismissed === 'true') {
        return
    }
    let decodedContext: string
    // File path of the workspace root directory
    const rootPath = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : null
    if (!rootPath) {
        return
    }
    // File path of the workspace recommendations file
    const filePath = vscode.Uri.file(`${rootPath}/.vscode/extensions.json`)
    // Create workspace edit API
    const newFile = new vscode.WorkspaceEdit()
    // Current state of the file: hasFile / noFile / null
    const fileState = await checkFileState()
    // Empty fileState means SG is already on the list
    // Skip CTA if SG is already on the list, and then store as dismissed so it won't show user again
    if (fileState === '') {
        await localStorageService.setValue(DISMISS_WORKSPACERECS_CTA_KEY, 'true')
        return
    }
    // Display Cta
    await vscode.window
        .showInformationMessage('Add Sourcegraph to your workspace recommendations?', 'Yes', 'No')
        .then(async answer => {
            if (answer === 'Yes') {
                switch (fileState) {
                    // Workspace recs file exists
                    case 'hasFile':
                        await addRec()
                        break
                    // Workspace recs does not exist
                    case 'noFile':
                        await createFile()
                        break
                    default:
                        return
                }
            }
            // Store as dismissed so it won't show user again when answer No or after being added
            await localStorageService.setValue(DISMISS_WORKSPACERECS_CTA_KEY, 'true')
        })
    // Check if the workspace recs file exists or not
    async function checkFileState(): Promise<string> {
        try {
            const bytes = await vscode.workspace.fs.readFile(filePath)
            const decoded = new TextDecoder('utf-8').decode(bytes)
            // Assume sourcegraph is in the recommendation list if 'sourcegraph.sourcegraph' exists in the file
            if (decoded.includes('sourcegraph.sourcegraph')) {
                return ''
            }
            // If sourcegraph not on the list, keep the decoded data for later
            decodedContext = decoded
            return 'hasFile'
        } catch {
            // This means we were not able to read the file because it doesn't exist
            return 'noFile'
        }
    }
    // Add Sourcegraph to the existing recommendations list
    async function addRec(): Promise<void> {
        if (!decodedContext.includes('recommendations')) {
            await createRec()
        }
        const recommendations = decodedContext
            .replace('"recommendations": [', '"recommendations": [ "sourcegraph.sourcegraph",')
            .replace('"recommendations":[', '"recommendations":["sourcegraph.sourcegraph",')
        const encodedData = new TextEncoder().encode(recommendations)
        await vscode.workspace.fs.writeFile(filePath, encodedData)
        // Reset value
        decodedContext = ''
        // Display a message to the user
        await vscode.window.showInformationMessage('Added Sourcegraph to your workspace recommendations successfully!')
    }
    // Create the file. It won't create one if it already exists
    // Then create rec list with sourcegraph in the list
    async function createFile(): Promise<void> {
        newFile.createFile(filePath, { ignoreIfExists: true })
        await createRec()
    }
    // create rec list with sourcegraph in the list
    async function createRec(): Promise<void> {
        const recommendations = '{"recommendations": ["sourcegraph.sourcegraph"]}'
        newFile.insert(filePath, new vscode.Position(0, 0), recommendations)
        await vscode.workspace.applyEdit(newFile)
        const textDocument = await vscode.workspace.openTextDocument(filePath)
        await textDocument.save()
        // Display a message to the user
        await vscode.window.showInformationMessage('Added Sourcegraph to your workspace recommendations successfully!')
    }
}
