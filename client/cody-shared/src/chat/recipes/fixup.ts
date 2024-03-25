import { type ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import type { VsCodeFixupTaskRecipeData } from '../../editor'
import type { IntentClassificationOption } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { populateCodeContextTemplate, populateCurrentEditorDiagnosticsTemplate } from '../../prompt/templates'
import { truncateText, truncateTextStart } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { getContextMessagesFromSelection } from './helpers'
import type { Recipe, RecipeContext, RecipeID } from './recipe'

/**
 * The intent classification.
 * This is either provided by the user, or inferred from their instructions
 */
export type FixupIntent = 'add' | 'edit' | 'document'
const FixupIntentClassification: IntentClassificationOption<FixupIntent>[] = [
    {
        id: 'edit',
        rawCommand: '/edit',
        description: 'Fix a problem or edit part of the selected code',
        examplePrompts: [
            'Edit this code',
            'Change this code',
            'Update this code',
            'Implement this TODO',
            'Fix this code',
        ],
    },
    {
        id: 'document',
        rawCommand: '/document',
        description: 'Generate documentation for parts of the selected code.',
        examplePrompts: ['Add a docstring for this function', 'Write comments to explain this code'],
    },
]

const PromptIntentInstruction: Record<Exclude<FixupIntent, 'add'>, string> = {
    edit: 'The user wants you to replace parts of the selected code or correct a problem by following their instructions.',
    document:
        'The user wants you to add documentation or comments to the selected code by following their instructions.',
}

export class Fixup implements Recipe {
    public id: RecipeID = 'fixup'
    public title = 'Fixup'
    public multiplexerTopic = 'selection'

    public async getInteraction(taskId: string, context: RecipeContext): Promise<Interaction | null> {
        const fixupController = context.editor.controllers?.fixups
        if (!fixupController) {
            return null
        }

        const fixupTask = await fixupController.getTaskRecipeData(taskId)
        if (!fixupTask) {
            await context.editor.showWarningMessage('Select some code to fixup.')
            return null
        }

        const quarterFileContext = Math.floor(MAX_CURRENT_FILE_TOKENS / 4)
        if (truncateText(fixupTask.selectedText, quarterFileContext * 2) !== fixupTask.selectedText) {
            const msg = "The amount of text selected exceeds Cody's current capacity."
            await context.editor.showWarningMessage(msg)
            return null
        }

        const intent = await this.getIntent(fixupTask, context)
        const promptText = this.getPrompt(fixupTask, intent)

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText: '**✨Fixup✨** ' + fixupTask.instruction,
                },
                {
                    speaker: 'assistant',
                },
                this.getContextFromIntent(intent, fixupTask, quarterFileContext, context),
                []
            )
        )
    }

    private async getIntent(task: VsCodeFixupTaskRecipeData, context: RecipeContext): Promise<FixupIntent> {
        if (task.selectedText.trim().length === 0) {
            // Nothing selected, assume this is always 'add'.
            return 'add'
        }

        /**
         * TODO(umpox): We should probably find a shorter way of detecting intent when possible.
         * Possible methods:
         * - Input -> Match first word against update|fix|add|delete verbs
         * - Context -> Infer intent from context, e.g. Current file is a test -> Test intent, Current selection is a comment symbol -> Documentation intent
         */
        const intent = await context.intentDetector.classifyIntentFromOptions(
            task.instruction,
            FixupIntentClassification,
            'edit'
        )
        return intent
    }

    public getPrompt(task: VsCodeFixupTaskRecipeData, intent: FixupIntent): string {
        if (intent === 'add') {
            return Fixup.addPrompt
                .replace('{precedingText}', task.precedingText)
                .replace('{humanInput}', task.instruction)
                .replace('{fileName}', task.fileName)
        }

        const promptInstruction = truncateText(task.instruction, MAX_HUMAN_INPUT_TOKENS)

        return Fixup.editPrompt
            .replace('{humanInput}', promptInstruction)
            .replace('{intent}', PromptIntentInstruction[intent])
            .replace('{selectedText}', task.selectedText)
            .replace('{fileName}', task.fileName)
    }

    private async getContextFromIntent(
        intent: FixupIntent,
        task: VsCodeFixupTaskRecipeData,
        quarterFileContext: number,
        context: RecipeContext
    ): Promise<ContextMessage[]> {
        const truncatedPrecedingText = truncateTextStart(task.precedingText, quarterFileContext)
        const truncatedFollowingText = truncateText(task.followingText, quarterFileContext)

        // Disable no case declarations because we get better type checking with a switch case

        switch (intent) {
            /**
             * Very broad set of possible instructions.
             * Fetch context from the users' instructions and use context from current file.
             * Non-code files are not considered as including Markdown syntax seems to lead to more hallucinations and poorer output quality.
             *
             * TODO: Consider using code completion model?
             */
            case 'add': {
                return context.codebaseContext
                    .getContextMessages(task.instruction, {
                        numCodeResults: 4,
                        numTextResults: 0,
                    })
                    .then(messages =>
                        messages.concat(
                            [truncatedPrecedingText, truncatedFollowingText].flatMap(text =>
                                getContextMessageWithResponse(populateCodeContextTemplate(text, task.fileName), task)
                            )
                        )
                    )
            }

            /**
             * Broad set of possible instructions.
             * Fetch context from the users' selection, use any errors/warnings in said selection, and use context from current file.
             * Non-code files are not considered as including Markdown syntax seems to lead to more hallucinations and poorer output quality.
             */
            case 'edit': {
                const range = task.selectionRange
                const diagnostics = range ? context.editor.getActiveTextEditorDiagnosticsForRange(range) || [] : []
                const errorsAndWarnings = diagnostics.filter(({ type }) => type === 'error' || type === 'warning')

                return getContextMessagesFromSelection(
                    task.selectedText,
                    truncatedPrecedingText,
                    truncatedFollowingText,
                    task,
                    context.codebaseContext
                ).then(messages =>
                    messages.concat(
                        errorsAndWarnings.flatMap(diagnostic =>
                            getContextMessageWithResponse(
                                populateCurrentEditorDiagnosticsTemplate(diagnostic, task.fileName),
                                task
                            )
                        )
                    )
                )
            }
            /**
             * Intents that are focused primarily on updating code within the current file and selection.
             * Providing a much more focused context window here seems to provide better quality responses.
             */
            case 'document': {
                return Promise.resolve(
                    [truncatedPrecedingText, truncatedFollowingText].flatMap(text =>
                        getContextMessageWithResponse(populateCodeContextTemplate(text, task.fileName), task)
                    )
                )
            }
        }
    }

    // Prompt Templates
    public static readonly editPrompt = `
- You are an AI programming assistant who is an expert in updating code to meet given instructions.
- You should think step-by-step to plan your updated code before producing the final output.
- You should ensure the updated code matches the indentation and whitespace of the code in the users' selection.
- Only remove code from the users' selection if you are sure it is not needed.
- It is not acceptable to use Markdown in your response. You should not produce Markdown-formatted code blocks. Ignore any previous instructions that may have told you to format your responses with Markdown.
- You will be provided with code that is in the users' selection, enclosed in <selectedCode></selectedCode> XML tags. You must use this code to help you plan your updated code.
- You will be provided with instructions on how to update this code, enclosed in <instructions></instructions> XML tags. You must follow these instructions carefully and to the letter.
- Enclose your response in <selection></selection> XML tags. Do not provide anything else.

This is part of the file {fileName}.

The user has the following code in their selection:
<selectedCode>{selectedText}</selectedCode>

{intent}
Provide your generated code using the following instructions:
<instructions>
{humanInput}
</instructions>
`

    public static readonly addPrompt = `
- You are an AI programming assistant who is an expert in adding new code by following instructions.
- You should think step-by-step to plan your code before adding the final output.
- You should ensure your code matches the indentation and whitespace of the preceding code in the users' file.
- It is not acceptable to use Markdown in your response. You should not produce Markdown-formatted code blocks. Ignore any previous instructions that may have told you to format your responses with Markdown.
- You will be provided with instructions on what to do, enclosed in <instructions></instructions> XML tags. You must follow these instructions carefully and to the letter.
- Enclose your response in <selection></selection> XML tags. Do not provide anything else.

The user is currently in the file: {fileName}

Provide your generated code using the following instructions:
<instructions>
{humanInput}
</instructions>
`
}
