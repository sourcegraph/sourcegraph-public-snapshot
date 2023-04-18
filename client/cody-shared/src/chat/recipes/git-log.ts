import { spawnSync } from 'child_process'

import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

const getPathFromWorkspaceRoot = (filePath: string, workspaceRootPath: string) =>
    filePath.replace(new RegExp(`^${workspaceRootPath}/`), '')

export class GitHistory implements Recipe {
    public getID(): string {
        return 'git-history'
    }

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return null
        }

        const activeTextEditor = context.editor.getActiveTextEditor()
        const currentFile = activeTextEditor && getPathFromWorkspaceRoot(activeTextEditor.filePath, dirPath)

        let path = ''
        const pathOptions = ['Repo', currentFile ? `Current file: ${currentFile}` : '', 'Custom path'].filter(Boolean)
        const selectedScope = await context.editor.showQuickPick(pathOptions)
        if (selectedScope?.startsWith('Current file')) {
            path = currentFile!
        } else if (selectedScope === 'Custom path') {
            const results = await context.editor.showOpenDialog({ canSelectFolders: true, canSelectFiles: true })
            if (!results) {
                return null
            }
            path = getPathFromWorkspaceRoot(results[0].path, dirPath)
        }

        const logFormat = '--pretty="Commit author: %an%nCommit message: %s%nChange description:%b%n"'
        const pathLabel = path ? `"${path}"` : 'my codebase'
        const items = [
            {
                label: 'Last 5 items',
                args: ['log', '-n5', logFormat],
                rawDisplayText: `What changed in ${pathLabel} in the last 5 commits?`,
            },
            {
                label: 'Last day',
                args: ['log', '--since', '1 day', logFormat],
                rawDisplayText: `What has changed in ${pathLabel} in the last day?`,
            },
            {
                label: 'Last week',
                args: ['log', "--since='1 week'", logFormat],
                rawDisplayText: `What changed in ${pathLabel} in the last week?`,
            },
        ]
        const selectedLabel = await context.editor.showQuickPick(items.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const selected = Object.fromEntries(
            items.map(({ label, args, rawDisplayText }) => [label, { args, rawDisplayText }])
        )[selectedLabel]

        const { args: gitArgs, rawDisplayText } = selected
        if (path) {
            gitArgs.push(path)
        }

        const gitLogCommand = spawnSync('git', ['--no-pager', ...gitArgs], { cwd: dirPath })
        const gitLogOutput = gitLogCommand.stdout.toString().trim()

        if (!gitLogOutput) {
            // TODO: Show the warning within the Chat UI.
            console.error(
                'No git history found for the selected option.',
                gitLogCommand.stderr.toString(),
                gitLogCommand.error?.message
            )
            return null
        }

        const truncatedGitLogOutput = truncateText(gitLogOutput, MAX_RECIPE_INPUT_TOKENS)
        if (truncatedGitLogOutput.length < gitLogOutput.length) {
            // TODO: Show the warning within the Chat UI.
            console.warn('Truncated extra long git log output, so summary may be incomplete.')
        }

        const promptMessage = `Summarize these commits:\n${truncatedGitLogOutput}\n\nProvide your response in the form of a bulleted list. Do not mention the commit hashes.`
        const assistantResponsePrefix = 'Here is a summary of recent changes:\n- '
        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText: rawDisplayText },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            Promise.resolve([])
        )
    }
}
