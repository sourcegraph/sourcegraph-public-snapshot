import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'

import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Recipes.module.css'

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
    'context-search': 'Codebase context search',
    'release-notes': 'Generate release notes',
    'pr-description': 'Generate pull request description',
}

export const Recipes: React.FunctionComponent<{ vscodeAPI: VSCodeWrapper }> = ({ vscodeAPI }) => {
    const onRecipeClick = (recipeID: RecipeID): void => {
        vscodeAPI.postMessage({ command: 'executeRecipe', recipe: recipeID })
    }

    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                <div className={styles.recipes}>
                    {Object.entries(recipesList).map(([key, value]) => (
                        <VSCodeButton
                            key={key}
                            className={styles.recipeButton}
                            type="button"
                            onClick={() => onRecipeClick(key as RecipeID)}
                        >
                            {value}
                        </VSCodeButton>
                    ))}
                </div>
            </div>
        </div>
    )
}
