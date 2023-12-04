interface SchemaCompatibleInferenceFormStep {
    root: string
    image: string
    commands: string[]
}

interface SchemaCompatibleInferenceFormJob {
    root: string
    indexer: string
    indexer_args: string[]
    requestedEnvVars: string[]
    local_steps: string[]
    outfile: string
    steps: SchemaCompatibleInferenceFormStep[]
}

export interface SchemaCompatibleInferenceFormData {
    index_jobs: SchemaCompatibleInferenceFormJob[]
}

/**
 * Form data with additional metadata that is unrelated to submission
 */
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

export interface InferenceFormJobStep extends Omit<SchemaCompatibleInferenceFormStep, 'commands'>, MetaIdentifier {
    commands: InferenceArrayValue[]
}

export interface InferenceFormJob
    extends MetaIdentifier,
        Omit<SchemaCompatibleInferenceFormJob, 'indexer_args' | 'requestedEnvVars' | 'local_steps' | 'steps'> {
    indexer_args: InferenceArrayValue[]
    requestedEnvVars: InferenceArrayValue[]
    local_steps: InferenceArrayValue[]
    steps: InferenceFormJobStep[]
}

export interface InferenceFormData {
    index_jobs: InferenceFormJob[]
    dirty: boolean
}
