import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { vscodeAPI } from './utils/VSCodeApi'

import './Recipes.css'

export const recipesList = {
    'explain-code-detailed': 'Explain selected code (detailed)',
    'explain-code-high-level': 'Explain selected code (high level)',
    'generate-unit-test': 'Generate a unit test',
    'generate-docstring': 'Generate a docstring',
    'improve-variable-names': 'Improve variable names',
    'translate-to-language': 'Translate to different language',
    'git-history': 'Summarize recent code changes',
    'find-code-smells': 'Smell code',
    fixup: 'Fixup code from inline instructions',
}

export function Recipes(): React.ReactElement {
    const onRecipeClick = (recipeID: string): void => {
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
