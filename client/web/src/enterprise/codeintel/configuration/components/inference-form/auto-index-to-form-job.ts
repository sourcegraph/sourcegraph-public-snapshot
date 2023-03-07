import { AutoIndexJobDescriptionFields, AutoIndexLsifPreIndexFields } from '../../../../../graphql-operations'
import { InferenceFormData, InferenceFormJobStep, InferenceFormJob } from './types'

const autoIndexStepToFormStep = (step: AutoIndexLsifPreIndexFields): InferenceFormJobStep => ({
    root: step.root,
    image: step.image ?? '',
    commands: step.commands,
})

const autoIndexJobToFormJob = (job: AutoIndexJobDescriptionFields): InferenceFormJob => ({
    root: job.root,
    indexer: job.indexer?.imageName ?? '',
    indexer_args: job.steps.index.indexerArgs,
    requestedEnvVars: job.steps.index.requestedEnvVars ?? [],
    local_steps: job.steps.index.commands,
    outfile: job.steps.index.outfile ?? '',
    steps: job.steps.preIndex.map(autoIndexStepToFormStep),
    meta: {
        comparisonKey: job.comparisonKey,
    },
})

export const autoIndexJobsToFormData = (jobs: AutoIndexJobDescriptionFields[]): InferenceFormData => ({
    index_jobs: jobs.map(autoIndexJobToFormJob),
})
