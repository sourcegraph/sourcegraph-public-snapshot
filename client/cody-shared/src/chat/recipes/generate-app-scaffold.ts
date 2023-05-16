import * as os from 'os'

import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'
import { ScaffoldGenerator } from './scaffold'

export class AppScaffold extends ScaffoldGenerator implements Recipe {
    public id: RecipeID = 'app-scaffold'

    public async getInteraction(_humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const dirPath = context.editor.getWorkspaceRootPath()
        if (!dirPath) {
            return Promise.resolve(null)
        }

        // get the root path where the app scaffold will be saved
        const rootPath = os.homedir()
        if (!rootPath) {
            return null
        }

        const rawDisplayText = 'Generating the scaffold of the selected app'

        const quickPickItems = [
            {
                label: 'React App',
            },
        ]

        const selectedLabel = await context.editor.showQuickPick(quickPickItems.map(e => e.label))
        if (!selectedLabel) {
            return null
        }
        const appName = await context.editor.showInputBox('Enter the app name')

        if (!appName) {
            const emptyAppNameMessage = 'No app name provided to generate scaffold'
            return new Interaction(
                { speaker: 'human', displayText: rawDisplayText },
                {
                    speaker: 'assistant',
                    prefix: emptyAppNameMessage,
                    text: emptyAppNameMessage,
                },
                Promise.resolve([])
            )
        }

        let generatorStatus = false
        let messsage = ''
        switch (selectedLabel) {
            case 'React App':
                // eslint-disable-next-line no-case-declarations
                const res = this.generateReactApp(rootPath, appName.trim())
                generatorStatus = res.generatorStatus
                messsage = res.generatorMessage
                break
            default:
                messsage = 'Generation unsuccessful'
        }

        let assistantResponsePrefix = 'The app scaffold has been generated successfully'

        // if app generation is unsucessful return the message, with the steps to fix error
        if (!generatorStatus) {
            assistantResponsePrefix = 'There were some errors while generating the selected app scaffold'
            const promptMessage = `You encountered this error messages ${messsage} while generating the ${selectedLabel}.\nSummarise the actual errors first.\nReturn a response for possible suggestioons to fix the error in stepwise format.`

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

        return new Interaction(
            { speaker: 'human', displayText: rawDisplayText },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            Promise.resolve([])
        )
    }
}
