import { spawnSync } from 'child_process'
import * as path from 'path'

import * as vscode from 'vscode'

import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class SummarizeChanges implements Recipe {
    public id = 'summarize-changes'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return null
        }

        // TODO: replace with context.editor
        const timeRangePicker = await vscode.window.createQuickPick()
        timeRangePicker.title = 'Over what time period would you like to summarize changes?'
        timeRangePicker.items = [{ label: 'last day' }, { label: 'last week' }, { label: 'since my last commit' }]
        timeRangePicker.onDidChangeValue(e => {
            console.log('# e', e)
        })
        await timeRangePicker.show()

        // Get scope
        const scopePicker = await vscode.window.createQuickPick()
        scopePicker.title = 'What set of code do you want to summarize changes over?'

        const filePath = context.editor.getActiveTextEditor()?.filePath
        if (filePath) {
            const dir = path.dirname(filePath)
            console.log(dir)
        }
        // set scope items to be the current directory, each parent/ancestor directory all the way up to the repository root

        return null
    }
}
