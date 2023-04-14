export interface IntentDetector {
    isCodebaseContextRequired(input: string): Promise<boolean | Error>
    isEditorContextRequired(input: string): boolean | Error
}
