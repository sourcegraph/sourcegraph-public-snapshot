export interface QueryInfo {
    needsCodebaseContext: boolean
    needsCurrentFileContext: boolean
}

export interface IntentDetector {
    detect(text: string): Promise<QueryInfo | Error>
}
