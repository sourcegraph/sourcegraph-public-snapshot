import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../graphql-operations'

export const kubernetes: RepogroupMetadata = {
    title: 'Kubernetes',
    name: 'kubernetes',
    url: '/kubernetes',
    description: 'Explore Kubernetes repositories on GitHub. Search with examples below.',
    examples: [
        {
            title:
                'Use a ReplicationController configuration to ensure specified number of pod replicas are running at any one time',
            query: 'file:pod.yaml content:"kind: ReplicationController"',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Look for outdated `apiVersions` of admission webhooks',
            description: `This apiVersion has been deprecated in favor of "admissionregistration.k8s.io/v1".
            You can read more about this at https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/`,
            query: 'content:"apiVersion: admissionregistration.k8s.io/v1beta1"',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find Prometheus usage in YAML files',
            query: 'lang:yaml prom/prometheus',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Search for examples of the sidecar pattern in Go',
            query: 'lang:go sidecar',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Browse diffs for recent code changes',
            query: 'type:diff after:"1 week ago"',
            patternType: SearchPatternType.literal,
        },
    ],
    homepageDescription: 'Search within the Kubernetes community.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/kubernetes.svg',
}
