import { spawnSync } from 'child_process'

import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'

export class ReleaseNotes implements Recipe {
    public id: RecipeID = 'release-notes'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return null
        }

        let quickPickItems = []
        const logFormat = '--pretty="Commit author: %an%nCommit message: %s%nChange description:%b%n"'

        // check for tags first
        const gitTagCommand = spawnSync('git', ['tag', '--sort=-creatordate'], { cwd: dirPath })
        const gitTagOutput = gitTagCommand.stdout.toString().trim()
        let tagsPromptText = ''

        if (gitTagOutput) {
            const tags = gitTagOutput.split(/\r?\n/)
            for (const tag of tags.slice(0, 3)) {
                quickPickItems.push({
                    label: tag,
                    args: ['log', tag, logFormat],
                })
            }
            tagsPromptText =
                'Do not include information about any other tags version number if any included in the commits.'
        } else {
            quickPickItems = [
                {
                    label: 'Last week',
                    args: ['log', "--since='1 week'", logFormat],
                },
                {
                    label: 'Last 2 weeks',
                    args: ['log', "--since='2 week'", logFormat],
                },
                {
                    label: 'Last 4 weeks',
                    args: ['log', "--since='4 week'", logFormat],
                },
            ]
        }

        const selectedLabel = await context.editor.showQuickPick(quickPickItems.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const selected = Object.fromEntries(quickPickItems.map(({ label, args }) => [label, { args }]))[selectedLabel]

        const { args: gitArgs } = selected

        const gitLogCommand = spawnSync('git', ['--no-pager', ...gitArgs], { cwd: dirPath })
        const gitLogOutput = gitLogCommand.stdout.toString().trim()
        const rawDisplayText = `Generate release notes for the changes made since ${selectedLabel}`

        if (!gitLogOutput) {
            const emptyGitLogMessage = 'No recent changes found to generate release notes.'
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
        console.log(truncatedGitLogOutput)
        let truncatedLogMessage = ''
        if (truncatedGitLogOutput.length < gitLogOutput.length) {
            truncatedLogMessage = 'Truncated extra long git log output, so release notes may miss some changes.'
        }

        const promptMessage = `Generate release notes by summarising these commits:\n${truncatedGitLogOutput}\n\nUse proper heading format for the release notes.\n\n${tagsPromptText}.Do not include other changes and dependency updates.`
        const assistantResponsePrefix = `Here is the generated release notes for ${selectedLabel}\n${truncatedLogMessage}`
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
