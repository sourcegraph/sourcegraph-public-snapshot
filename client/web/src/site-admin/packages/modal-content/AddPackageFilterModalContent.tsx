import { useState } from 'react'

import { useMutation } from '@sourcegraph/http-client'
import { ErrorAlert, PageHeader } from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../../components/FilteredConnection'
import {
    AddPackageRepoFilterResult,
    AddPackageRepoFilterVariables,
    PackageMatchBehaviour,
    SiteAdminPackageFields,
} from '../../../graphql-operations'
import { addPackageRepoFilterMutation } from '../backend'
import { BehaviourSelect } from '../components/BehaviourSelect'
import { MultiPackageForm } from '../components/MultiPackageForm'
import { SinglePackageForm } from '../components/SinglePackageForm'

import styles from './AddPackageFilterModalContent.module.scss'

export interface AddPackageFilterModalContentProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    onDismiss: () => void
}

export type BlockType = 'single' | 'multiple'

export const AddPackageFilterModalContent: React.FunctionComponent<AddPackageFilterModalContentProps> = ({
    node,
    filters,
    onDismiss,
}) => {
    const [behaviour, setBehaviour] = useState<PackageMatchBehaviour>(PackageMatchBehaviour.BLOCK)
    const [blockType, setBlockType] = useState<BlockType>('single')

    const [addPackageRepoFilter, { error }] = useMutation<AddPackageRepoFilterResult, AddPackageRepoFilterVariables>(
        addPackageRepoFilterMutation,
        { onCompleted: onDismiss }
    )

    return (
        <>
            <PageHeader path={[{ text: 'Add package filter' }]} headingElement="h2" className={styles.header} />
            <div className={styles.content}>
                <BehaviourSelect value={behaviour} onChange={setBehaviour} />
                {blockType === 'single' ? (
                    <SinglePackageForm
                        node={node}
                        filters={filters}
                        setType={setBlockType}
                        onDismiss={onDismiss}
                        onSave={blockState =>
                            addPackageRepoFilter({
                                variables: {
                                    behaviour,
                                    kind: blockState.ecosystem,
                                    filter: {
                                        versionFilter: {
                                            packageName: blockState.name,
                                            versionGlob: blockState.versionFilter,
                                        },
                                    },
                                },
                            })
                        }
                    />
                ) : (
                    <MultiPackageForm
                        node={node}
                        filters={filters}
                        setType={setBlockType}
                        onDismiss={onDismiss}
                        onSave={blockState =>
                            addPackageRepoFilter({
                                variables: {
                                    behaviour,
                                    kind: blockState.ecosystem,
                                    filter: {
                                        nameFilter: {
                                            packageGlob: blockState.nameFilter,
                                        },
                                    },
                                },
                            })
                        }
                    />
                )}
                {error && <ErrorAlert error={error} />}
            </div>
        </>
    )
}
