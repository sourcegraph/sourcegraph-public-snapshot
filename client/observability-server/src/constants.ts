import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'

export const SDK_INFO = {
    [SemanticResourceAttributes.TELEMETRY_SDK_NAME]: 'libhoney',
    [SemanticResourceAttributes.TELEMETRY_SDK_LANGUAGE]: 'node.js',
}

export const BUILDKITE_INFO = {
    'buildkite.build.url': process.env.BUILDKITE_BUILD_URL,
    'buildkite.build.id': process.env.BUILDKITE_BUILD_ID,
    'buildkite.step.key': process.env.BUILDKITE_STEP_KEY,
    'buildkite.step.id': process.env.BUILDKITE_STEP_ID,
}
