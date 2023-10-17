import { useCallback } from 'react'

import { mdiDelete, mdiPencil } from '@mdi/js'
import classNames from 'classnames'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import { LoadingSpinner, ErrorAlert, Badge, Input, Label, Button, Icon, Alert } from '@sourcegraph/wildcard'

import {
    PackageMatchBehaviour,
    type PackageRepoFilterFields,
    type PackageRepoFiltersResult,
    type PackageRepoFiltersVariables,
} from '../../../graphql-operations'
import { packageRepoFiltersQuery, deletePackageRepoFilterMutation } from '../backend'
import { PackageExternalServiceMap } from '../constants'

import styles from './ManagePackageFiltersModalContent.module.scss'

export interface ManagePackageFiltersModalContentProps {
    setActiveFilter: (filter: PackageRepoFilterFields) => void
    onDismiss: () => void
}

export const ManagePackageFiltersModalContent: React.FunctionComponent<ManagePackageFiltersModalContentProps> = ({
    setActiveFilter,
    onDismiss,
}) => {
    const { data, loading, error } = useQuery<PackageRepoFiltersResult, PackageRepoFiltersVariables>(
        packageRepoFiltersQuery,
        {}
    )

    return (
        <>
            {loading || !data ? (
                <LoadingSpinner className="d-block mx-auto mt-3" />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : (
                <div className={styles.content}>
                    <div className={styles.grid}>
                        <Label className={styles.label}>Behavior</Label>
                        <Label className={styles.label}>Ecosystem</Label>
                        <Label className={styles.label}>Package filter</Label>
                        <Label className={styles.label}>Version filter</Label>
                    </div>
                    {(data.packageRepoFilters ?? []).length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {(data.packageRepoFilters ?? []).map(filter => (
                                <PackageFilterNode key={filter.id} node={filter} setActiveFilter={setActiveFilter} />
                            ))}
                        </ul>
                    ) : (
                        <Alert variant="info" className="mt-3">
                            No package filters found
                        </Alert>
                    )}
                </div>
            )}
            <div className={styles.closeAction}>
                <Button variant="secondary" onClick={onDismiss} className="mt-2">
                    Close
                </Button>
            </div>
        </>
    )
}

interface PackageFilterNodeProps {
    node: PackageRepoFilterFields
    setActiveFilter: (filter: PackageRepoFilterFields) => void
}

const PackageFilterNode: React.FunctionComponent<PackageFilterNodeProps> = ({ node, setActiveFilter }) => {
    const nameValue = node.versionFilter?.packageName || node.nameFilter?.packageGlob
    const versionValue = node.versionFilter?.versionGlob || null

    const [deletePackageFilter, { error }] = useMutation(deletePackageRepoFilterMutation)

    const onDelete = useCallback(() => {
        if (window.confirm('Are you sure you want to delete this filter?')) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            deletePackageFilter({
                variables: { id: node.id },
                update: cache => {
                    cache.modify({
                        fields: {
                            packageRepoFilters(
                                packageFilters: PackageRepoFiltersResult['packageRepoFilters'],
                                { readField }
                            ) {
                                return packageFilters?.filter(filter => readField('id', filter) !== node.id)
                            },
                        },
                    })
                },
            })
        }
    }, [deletePackageFilter, node.id])

    return (
        <li className={classNames('list-group-item', styles.item)}>
            <div className={styles.grid}>
                <Badge className={styles.badge} variant="outlineSecondary">
                    {node.behaviour === PackageMatchBehaviour.ALLOW ? 'Allowlist' : 'Blocklist'}
                </Badge>
                <Badge className={styles.badge} variant="outlineSecondary">
                    {PackageExternalServiceMap[node.kind]?.label ?? node.kind}
                </Badge>
                <Input value={nameValue} readOnly={true} className={styles.input} />
                {versionValue && (
                    <Input value={versionValue} readOnly={true} className={classNames(styles.input, 'ml-2')} />
                )}
                <div className={styles.actions}>
                    <Button
                        onClick={() => setActiveFilter(node)}
                        variant="icon"
                        aria-label="Edit"
                        className="text-primary mr-3"
                    >
                        <Icon svgPath={mdiPencil} aria-hidden={true} />
                    </Button>
                    <Button onClick={onDelete} variant="icon" aria-label="Delete" className="text-danger">
                        <Icon svgPath={mdiDelete} aria-hidden={true} />
                    </Button>
                </div>
            </div>
            {error && <ErrorAlert error={error} className="mt-2" />}
        </li>
    )
}
