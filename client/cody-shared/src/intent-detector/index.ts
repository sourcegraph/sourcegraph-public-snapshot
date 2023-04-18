export interface IntentDetector {
    isCodebaseContextRequired(input: string): Promise<boolean | Error>
    isEditorContextRequired(input: string): Promise<boolean | Error>
// TODO: Add this:   isEditorBroaderFileContextRequired(input: string): Promise<boolean | Error>
}
