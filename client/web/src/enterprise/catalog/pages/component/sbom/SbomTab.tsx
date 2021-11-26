/* eslint-disable unicorn/prevent-abbreviations */
import classNames from 'classnames'
import React, { useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SourceLocationSetSBOMResult, SourceLocationSetSBOMVariables } from '../../../../../graphql-operations'

import { BOM, BOMRef } from './cyclonedx'

interface Props {
    sourceLocationSet: Scalars['ID']
    className?: string
}

const SOURCE_LOCATION_SET_SBOM = gql`
    query SourceLocationSetSBOM($node: ID!) {
        node(id: $node) {
            ...SourceLocationSetSBOMFields
        }
    }

    fragment SourceLocationSetSBOMFields on SourceLocationSet {
        __typename
        id
        cyclonedx
    }
`

export const SbomTab: React.FunctionComponent<Props> = ({ sourceLocationSet, className }) => {
    const { data, error, loading } = useQuery<SourceLocationSetSBOMResult, SourceLocationSetSBOMVariables>(
        SOURCE_LOCATION_SET_SBOM,
        {
            variables: { node: sourceLocationSet },
            fetchPolicy: 'cache-first',
        }
    )

    if (loading && !data) {
        return <LoadingSpinner />
    }
    if (error && !data) {
        return <ErrorAlert error={error} />
    }
    if (!data || !data.node || !('cyclonedx' in data.node)) {
        return <ErrorAlert error="TODO(sqs)" />
    }
    if (!data.node.cyclonedx) {
        return <ErrorAlert error="No SBOM information" />
    }

    return (
        <div className={classNames('row no-gutters', className)}>
            <Cyclonedx data={data.node.cyclonedx} />
        </div>
    )
}

const Cyclonedx: React.FunctionComponent<{ data: string }> = ({ data }) => {
    const cdx: BOM = useMemo(() => JSON.parse(data), [data])

    const subjectRef = cdx.metadata?.component?.['bom-ref']

    const directDeps: BOMRef[] | undefined = useMemo(() => {
        if (cdx.dependencies) {
            for (const dep of cdx.dependencies) {
                if (dep.ref === subjectRef) {
                    return dep.dependsOn
                }
            }
        }
        return undefined
    }, [cdx.dependencies, subjectRef])

    return (
        <pre>
            <code>{directDeps?.join('\n')}</code>
        </pre>
    )
}
