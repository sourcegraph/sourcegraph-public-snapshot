import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const kubernetes: RepogroupMetadata = {
    title: 'Kubernetes',
    name: 'kubernetes',
    url: '/kubernetes',
    description: 'Explore Kubernetes repositories on GitHub. Search with examples below.',
    examples: [
        {
            title:
                'Use a ReplicationController configuration to ensure specified number of pod replicas are running at any one time',
            exampleQuery: (
                <>
                    <span className="web-content__link">file:</span>pod.yaml{' '}
                    <span className="web-content__link">content:</span>"kind: ReplicationController"
                </>
            ),
            rawQuery: 'file:pod.yaml content:"kind: ReplicationController"',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Look for outdated `apiVersions` of admission webhooks',
            exampleQuery: (
                <>
                    <span className="web-content__link">content:</span>"apiVersion:
                    admissionregistration.k8s.io/v1beta1"
                </>
            ),
            description: `This apiVersion has been deprecated in favor of "admissionregistration.k8s.io/v1".
            You can read more about this at https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/`,
            rawQuery: 'content:"apiVersion: admissionregistration.k8s.io/v1beta1"',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find Prometheus usage in YAML files',
            exampleQuery: (
                <>
                    <span className="web-content__link">lang:</span>yaml prom/prometheus
                </>
            ),
            rawQuery: 'lang:yaml prom/prometheus',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Search for examples of the sidecar pattern in Go',
            exampleQuery: (
                <>
                    <span className="web-content__link">lang:</span>go sidecar
                </>
            ),
            rawQuery: 'lang:go sidecar',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Browse diffs for recent code changes',
            exampleQuery: (
                <>
                    <span className="web-content__link">type:</span>diff{' '}
                    <span className="web-content__link">after:</span>"1 week ago"
                </>
            ),
            rawQuery: 'type:diff after:"1 week ago"',
            patternType: SearchPatternType.literal,
        },
    ],
    homepageDescription: 'Search within the Kubernetes community.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/kubernetes.svg',
}
