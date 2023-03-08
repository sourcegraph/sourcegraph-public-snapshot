interface MetaIdentifier {
    meta: {
        id: string
    }
}

// InferenceJobs only return a unique ID for the actual job, not values within the job.
// As we want to build a dynamic form, we need each array of values to have a unique id.
export interface InferenceArrayValue extends MetaIdentifier {
    value: string
}

export interface InferenceFormJobStep extends MetaIdentifier {
    root: string
    image: string
    commands: InferenceArrayValue[]
}

export interface InferenceFormJob extends MetaIdentifier {
    root: string
    indexer: string
    indexer_args: InferenceArrayValue[]
    requestedEnvVars: InferenceArrayValue[]
    local_steps: InferenceArrayValue[]
    outfile: string
    steps: InferenceFormJobStep[]
}

export interface InferenceFormData {
    index_jobs: InferenceFormJob[]
}
