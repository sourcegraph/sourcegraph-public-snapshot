import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { vscodeAPI } from './utils/VSCodeApi'

import './Recipes.css'

import { EventLogger } from './eventLogger'

// eventLogger is used to log events
const eventLogger = new EventLogger('https://sourcegraph.com/')

export const recipesList = {
    'explain-code-detailed': 'Explain selected code (detailed)',
    'explain-code-high-level': 'Explain selected code (high level)',
    'generate-unit-test': 'Generate a unit test',
    'generate-docstring': 'Generate a docstring',
    'improve-variable-names': 'Improve variable names',
    'translate-to-language': 'Translate to different language',
    'git-history': 'Summarize recent code changes',
}

export function Recipes(): React.ReactElement {
    const onRecipeClick = (recipeID: string): void => {
        // log event
        try {
            eventLogger.logEvent({
                event: 'CodyVSCodeExtenstion:recipe:" + recipeID + ":clicked',
            })
        } catch (error) {
            console.log(error)
        }
        vscodeAPI.postMessage({ command: 'executeRecipe', recipe: recipeID })
    }

    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                <div className="recipes">
                    {Object.entries(recipesList).map(([key, value]) => (
                        <VSCodeButton
                            key={key}
                            className="recipe-button"
                            type="button"
                            onClick={() => onRecipeClick(key)}
                        >
                            {value}
                        </VSCodeButton>
                    ))}
                </div>
            </div>
        </div>
    )
}
