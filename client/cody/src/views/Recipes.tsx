import './App.css'
import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { MessageFromWebview, vscodeAPI } from './utils/vscodeAPI'

export const recipesList = {
    explainCode: 'Explain selected code (detailed)',
    explainCodeHighLevel: 'Explain selected code (high level)',
    generateUnitTest: 'Generate a unit test',
    generateDocstring: 'Generate a docstring',
    translateToLanguage: 'Translate to different language',
    gitHistory: 'Recent history',
}

function Recipes() {
    const onRecipeClick = (recipeID: string) => {
        vscodeAPI.postMessage({
            command: 'executeRecipe',
            recipe: recipeID,
        } as MessageFromWebview)
    }

    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                <div className="container-recipes">
                    {Object.entries(recipesList).map(([key, value]) => (
                        <VSCodeButton className="btn-recipe" type="button" onClick={() => onRecipeClick(key)}>
                            {value}
                        </VSCodeButton>
                    ))}
                </div>
            </div>
        </div>
    )
}

export default Recipes
