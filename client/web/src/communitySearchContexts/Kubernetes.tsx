import React from 'react'

import { SearchPatternType } from '../graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const kubernetes: CommunitySearchContextMetadata = {
    title: 'Kubernetes',
    spec: 'kubernetes',
    url: '/kubernetes',
    description: 'Explore Kubernetes repositories on GitHub. Search with examples below.',
    examples: [
        {
            title: 'Use a ReplicationController configuration to ensure specified number of pod replicas are running at any one time',
            query: 'file:pod.yaml content:"kind: ReplicationController"',
            patternType: SearchPatternType.standard,
        },
        {
            title: 'Look for outdated `apiVersions` of admission webhooks',
            description: `This apiVersion has been deprecated in favor of "admissionregistration.k8s.io/v1".
            You can read more about this at https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/`,
            query: 'content:"apiVersion: admissionregistration.k8s.io/v1beta1"',
            patternType: SearchPatternType.standard,
        },
        {
            title: 'Find Prometheus usage in YAML files',
            query: 'lang:yaml prom/prometheus',
            patternType: SearchPatternType.standard,
        },
        {
            title: 'Search for examples of the sidecar pattern in Go',
            query: 'lang:go sidecar',
            patternType: SearchPatternType.standard,
        },
        {
            title: 'Browse diffs for recent code changes',
            query: 'type:diff after:"1 week ago"',
            patternType: SearchPatternType.standard,
        },
    ],
    homepageDescription: 'Search within the Kubernetes community.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/kubernetes.svg',
}

export const KubernetesCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={kubernetes} />
