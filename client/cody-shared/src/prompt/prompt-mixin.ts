import { InteractionMessage } from '../chat/transcript/messages'

// Prompt mixins elaborate every prompt presented to the LLM. Add a prompt mixin to prompt for cross-cutting concerns relevant to multiple recipes.
export class PromptMixin {
    private static mixins_: PromptMixin[] = []

    // Adds a prompt mixin to the global set.
    public static add(mixin: PromptMixin): void {
        this.mixins_.push(mixin)
    }

    // Prepends all of the mixins to `humanMessage`. Modifies and returns `humanMessage`.
    public static mixInto(humanMessage: InteractionMessage): InteractionMessage {
        const mixins = this.mixins_.map(mixin => mixin.prompt).join('\n\n')
        if (mixins) {
            // Stuff the prompt mixins at the start of the human text.
            // Note we do not reflect them in displayText.
            return { ...humanMessage, text: `${mixins}\n\n${humanMessage.text}` }
        }
        return humanMessage
    }

    // Creates a mixin with the given, fixed prompt to insert.
    constructor(private readonly prompt: string) {}
}

// Creates a prompt mixin to get Cody to reply in the given language, for example "en-AU" for "Australian English".
export function languagePromptMixin(languageCode: string): PromptMixin {
    return new PromptMixin(
        `Unless instructed otherwise, reply in the language with RFC5646/ISO language code "${languageCode}".`
    )
}
