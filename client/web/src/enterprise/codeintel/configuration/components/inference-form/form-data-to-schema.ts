import type { InferenceFormData, SchemaCompatibleInferenceFormData } from './types'

export const formDataToSchema = (formData: InferenceFormData): SchemaCompatibleInferenceFormData => {
    // Remove all meta information from the form data
    const cleanJobs = formData.index_jobs.map(job => {
        const { meta, ...cleanJob } = job
        return {
            ...cleanJob,
            steps: job.steps.map(step => {
                const { meta, ...cleanStep } = step
                return {
                    ...cleanStep,
                    commands: step.commands.map(command => command.value),
                }
            }),
            indexer_args: job.indexer_args.map(arg => arg.value),
            requestedEnvVars: job.requestedEnvVars.map(envVar => envVar.value),
            local_steps: job.local_steps.map(step => step.value),
        }
    })

    return {
        index_jobs: cleanJobs,
    }
}
