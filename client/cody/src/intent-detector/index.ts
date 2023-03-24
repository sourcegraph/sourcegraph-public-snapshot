export interface IntentDetector {
    isCodebaseContextRequired(input: string): Promise<boolean | Error>
}
