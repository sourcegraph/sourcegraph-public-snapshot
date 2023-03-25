import { spawnSync } from 'child_process'
import * as path from 'path'

import { MAX_RECIPE_INPUT_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { truncateText } from '@sourcegraph/cody-shared/src/prompt/truncation'
import { getShortTimestamp } from '@sourcegraph/cody-shared/src/timestamp'

import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { renderMarkdown } from '../markdown'
import { Interaction } from '../transcript/interaction'

import { Recipe } from './recipe'

export class GitHistory implements Recipe {
    public getID(): string {
        return 'git-history'
    }

    public async getInteraction(
        _humanChatInput: string,
        editor: Editor,
        _intentDetector: IntentDetector,
        _codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const activeEditor = editor.getActiveTextEditor()
        if (!activeEditor) {
            return null
        }
        const dirPath = path.dirname(activeEditor.filePath)

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
        const selectedLabel = await editor.showQuickPick(items.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const selected = Object.fromEntries(
            items.map(({ label, args, rawDisplayText }) => [label, { args, rawDisplayText }])
        )[selectedLabel]

        const { args: gitArgs, rawDisplayText } = selected

        const gitLogCommand = spawnSync('git', ['--no-pager', ...gitArgs], { cwd: dirPath })
        const gitLogOutput = gitLogCommand.stdout.toString().trim()

        if (!gitLogOutput) {
            // TODO: Show the warning within the Chat UI.
            // editor.showWarningMessage('No git history found for the selected option.')
            return null
        }

        const truncatedGitLogOutput = truncateText(gitLogOutput, MAX_RECIPE_INPUT_TOKENS)
        if (truncatedGitLogOutput.length < gitLogOutput.length) {
            // TODO: Show the warning within the Chat UI.
            // editor.showWarningMessage('Truncated extra long git log output, so summary may be incomplete.')
        }

        const timestamp = getShortTimestamp()
        const promptMessage = `Summarize these commits:\n${truncatedGitLogOutput}\n\nProvide your response in the form of a bulleted list. Do not mention the commit hashes.`
        const assistantResponsePrefix = 'Here is a summary of recent changes:\n- '
        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText: renderMarkdown(rawDisplayText), timestamp },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
                displayText: '',
                timestamp,
            },
            Promise.resolve([])
        )
    }
}
