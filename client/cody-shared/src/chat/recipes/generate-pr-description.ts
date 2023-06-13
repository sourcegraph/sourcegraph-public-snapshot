import { spawnSync } from 'child_process'
import { readFileSync } from 'fs'
import path from 'path'

import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'

export class PrDescription implements Recipe {
    public id: RecipeID = 'pr-description'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return Promise.resolve(null)
        }

        const logFormat = '--pretty="Commit author: %an%nCommit message: %s%nChange description:%b%n"'

        const rawDisplayText = 'Generating the PR description'

        const templateFormatArgs = [
            'pull_request_template.md',
            'PULL_REQUEST_TEMPLATE.md',
            'docs/PULL_REQUEST_TEMPLATE.md',
            'docs/pull_request_template.md',
            '.github/pull_request_template.md',
            '.github/PULL_REQUEST_TEMPLATE.md',
        ]

        const checkPrTemplate = spawnSync('git', ['ls-files', ...templateFormatArgs], { cwd: dirPath })
        const prTemplateOutput = checkPrTemplate.stdout.toString().trim()

        let prTemplateContent = ''

        if (prTemplateOutput) {
            const templatePath = path.join(dirPath.trim(), prTemplateOutput)
            prTemplateContent = readFileSync(templatePath).toString()
        }

        const gitCommit = spawnSync('git', ['log', 'origin/HEAD..HEAD', logFormat], { cwd: dirPath })
        const gitCommitOutput = gitCommit.stdout.toString().trim()

        if (!gitCommitOutput) {
            const emptyGitCommitMessage = 'No commits history found in the current branch.'
            return new Interaction(
                { speaker: 'human', displayText: rawDisplayText },
                {
                    speaker: 'assistant',
                    prefix: emptyGitCommitMessage,
                    text: emptyGitCommitMessage,
                },
                Promise.resolve([]),
                []
            )
        }

        const truncatedGitCommitOutput = truncateText(gitCommitOutput, MAX_RECIPE_INPUT_TOKENS)
        let truncatedCommitMessage = ''
        if (truncatedGitCommitOutput.length < gitCommitOutput.length) {
            truncatedCommitMessage = 'Truncated extra long git log output, so PR description may be incomplete.'
        }

        const promptMessage = `Summarise these changes:\n${gitCommitOutput}\n\n made while working in the current git branch.\nUse this pull request template to ${prTemplateContent} generate a pull request description based on the committed changes.\nIf the PR template mentions a requirement to check the contribution guidelines, then just summarise the changes in bulletin format.\n If it mentions a test plan for the changes use N/A\n.`
        const assistantResponsePrefix = `Here is the PR description for the work done in your current branch:\n${truncatedCommitMessage}`
        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText: rawDisplayText },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            Promise.resolve([]),
            []
        )
    }
}
