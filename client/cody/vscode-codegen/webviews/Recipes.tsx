import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { WebviewMessage, vscodeAPI } from './utils/VSCodeApi'

import './Recipes.css'

export const recipesList = {
    explainCode: 'Explain selected code (detailed)',
    explainCodeHighLevel: 'Explain selected code (high level)',
    generateUnitTest: 'Generate a unit test',
    generateDocstring: 'Generate a docstring',
    translateToLanguage: 'Translate to different language',
    gitHistory: 'Recent history',
}

export function Recipes(): React.ReactElement {
    const onRecipeClick = (recipeID: string) => {
        vscodeAPI.postMessage({ command: 'executeRecipe', recipe: recipeID } as WebviewMessage)
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
