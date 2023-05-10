import { spawnSync } from 'child_process'
import { readFileSync } from 'fs'
import path from 'path'

import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class PrDescription implements Recipe {
    public id = 'pr-description'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return null
        }

        const logFormat = '--pretty="Commit author: %an%nCommit message: %s%nChange description:%b%n"'
        const items = [
            {
                label: 'Last 5 items',
                args: ['log', '-n5', logFormat],
                rawDisplayText: 'What changed in my codebase in the last 5 commits?',
            },
            {
                label: 'Last day',
                args: ['log', '--since', '1 day', logFormat],
                rawDisplayText: 'What has changed in my codebase in the last day?',
            },
            {
                label: 'Last week',
                args: ['log', "--since='1 week'", logFormat],
                rawDisplayText: 'What changed in my codebase in the last week?',
            },
        ]
        const selectedLabel = await context.editor.showQuickPick(items.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const selected = Object.fromEntries(
            items.map(({ label, args, rawDisplayText }) => [label, { args, rawDisplayText }])
        )[selectedLabel]

        const rawDisplayText = 'Generating the PR description'

        const templateFormatArgs = [
            'pull_request_template.md',
            'docs/pull_request_template.md',
            '.github/pull_request_template.md',
            '.github/PULL_REQUEST_TEMPLATE.md',
        ]

        const gitLogCommand = spawnSync('git', ['ls-files', ...templateFormatArgs], { cwd: dirPath })
        const gitLogOutput = gitLogCommand.stdout.toString().trim()

        const templatePath = path.join(dirPath.trim(), gitLogOutput)
        const fileContents = readFileSync(templatePath)
        console.log(fileContents)

        if (!gitLogOutput) {
            const emptyGitLogMessage = 'No recent changes found'
            return new Interaction(
                { speaker: 'human', displayText: rawDisplayText },
                {
                    speaker: 'assistant',
                    prefix: emptyGitLogMessage,
                    text: emptyGitLogMessage,
                },
                Promise.resolve([])
            )
        }

        const truncatedGitLogOutput = truncateText(gitLogOutput, MAX_RECIPE_INPUT_TOKENS)
        let truncatedLogMessage = ''
        if (truncatedGitLogOutput.length < gitLogOutput.length) {
            truncatedLogMessage = 'Truncated extra long git log output, so summary may be incomplete.'
        }

        const promptMessage = `Summarize these commits:\n${truncatedGitLogOutput}\n\nProvide your response in the form of a bulleted list. Do not mention the commit hashes.`
        const assistantResponsePrefix = `Here is a summary of recent changes:\n${truncatedLogMessage}`
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
