export interface InferenceFormJobStep {
    root: string
    image: string
    commands: string[]
}

export interface InferenceFormJob {
    root: string
    indexer: string
    indexer_args: string[]
    requestedEnvVars: string[] | null
    local_steps: string[]
    outfile: string
    steps: InferenceFormJobStep[]

    // Data that used for something else than form submission
    meta: {
        comparisonKey: string
    }
}

export interface InferenceFormData {
    index_jobs: InferenceFormJob[]
}
