import { ContextSearchOptions, Message } from '../../types';

export interface RecipePrompt {
  displayText: string;
  contextMessages: Message[];
  promptMessage: Message;
  botResponsePrefix: string;
}

export interface Recipe {
  getID(): string;
  getPrompt(
    maxTokens: number,
    getEmbeddingsContextMessages: (
      query: string,
      options: ContextSearchOptions
    ) => Promise<Message[]>
  ): Promise<RecipePrompt | null>;
}
