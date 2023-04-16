import { Message } from '../../sourcegraph-api'

export interface ChatMessage extends Message, MessageActions {
    displayText?: string
}

/** Actions performed by the assistant in sending a message. */
export interface AssistantActions {
    /** The files that the assistant read. */
    contextFiles?: string[]

    tmpDescription?: string
}

/** Actions performed by the human in sending a message. */
export interface HumanActions {
    /**
     * A description of the recipe that the human ran.
     */
    recipeDescription?: string
}

export interface InteractionMessage extends Message, MessageActions {
    displayText?: string
    prefix?: string
}

export interface MessageActions {
    /**
     * Actions performed by the assistant (such as searching for and reading files).
     *
     * Only set for messages sent by the assistant.
     */
    assistantActions?: AssistantActions

    /**
     * Actions performed by the human (such as running a recipe).
     *
     * Only set for messages sent by the assistant.
     */
    humanActions?: HumanActions
}

export interface UserLocalHistory {
    chat: ChatHistory
    input: string[]
}

export interface ChatHistory {
    [chatID: string]: ChatMessage[]
}
