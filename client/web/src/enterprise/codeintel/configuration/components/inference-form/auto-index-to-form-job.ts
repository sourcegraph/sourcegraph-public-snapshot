import { uniqueId } from 'lodash'

import type { AutoIndexJobDescriptionFields, AutoIndexLsifPreIndexFields } from '../../../../../graphql-operations'

import type { InferenceFormData, InferenceFormJobStep, InferenceFormJob } from './types'

const autoIndexStepToFormStep = (step: AutoIndexLsifPreIndexFields): InferenceFormJobStep => ({
    root: step.root,
    image: step.image ?? '',
    commands: step.commands.map(arg => ({
        value: arg,
        meta: {
            id: uniqueId(),
        },
    })),
    meta: {
        id: uniqueId(),
    },
})

const autoIndexJobToFormJob = (job: AutoIndexJobDescriptionFields): InferenceFormJob => ({
    root: job.root,
    indexer: job.indexer?.imageName ?? '',
    indexer_args: job.steps.index.indexerArgs.map(arg => ({
        value: arg,
        meta: {
            id: uniqueId(),
        },
    })),
    requestedEnvVars: (job.steps.index.requestedEnvVars ?? []).map(envVar => ({
        value: envVar,
        meta: {
            id: uniqueId(),
        },
    })),
    local_steps: job.steps.index.commands.map(command => ({
        value: command,
        meta: {
            id: uniqueId(),
        },
    })),
    outfile: job.steps.index.outfile ?? '',
    steps: job.steps.preIndex.map(autoIndexStepToFormStep),
    meta: {
        id: job.comparisonKey,
    },
})

interface AutoIndexJobsToFormDataParams {
    jobs: AutoIndexJobDescriptionFields[]
    dirty?: boolean
}
export const autoIndexJobsToFormData = ({ jobs, dirty = false }: AutoIndexJobsToFormDataParams): InferenceFormData => ({
    index_jobs: jobs.map(autoIndexJobToFormJob),
    dirty,
})
