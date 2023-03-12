import { useState } from 'react'

import { Modal } from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import { PackageRepoFilterFields } from '../../graphql-operations'

import {
    AddPackageFilterModalContent,
    AddPackageFilterModalContentProps,
} from './modal-content/AddPackageFilterModalContent'
import { EditPackageFilterModalContent } from './modal-content/EditPackageFilterModalContent'
import {
    ManagePackageFiltersModalContent,
    ManagePackageFiltersModalContentProps,
} from './modal-content/ManagePackageFiltersModalContent'

import styles from './PackagesModal.module.scss'

interface BaseProps {
    onDismiss: () => void
    filters: FilteredConnectionFilterValue[]
}

interface ManageFiltersProps extends BaseProps, Omit<ManagePackageFiltersModalContentProps, 'setActiveFilter'> {
    type: 'manage'
}

interface AddFilterProps extends BaseProps, AddPackageFilterModalContentProps {
    type: 'add'
}

type PackagesModalProps = ManageFiltersProps | AddFilterProps

export const PackagesModal: React.FunctionComponent<PackagesModalProps> = props => {
    const [activeFilter, setActiveFilter] = useState<PackageRepoFilterFields>()

    return (
        <Modal aria-label="Manage package filters" onDismiss={props.onDismiss} className={styles.modal}>
            {activeFilter ? (
                <EditPackageFilterModalContent
                    packageFilter={activeFilter}
                    filters={props.filters}
                    onDismiss={props.onDismiss}
                />
            ) : props.type === 'add' ? (
                <AddPackageFilterModalContent node={props.node} filters={props.filters} onDismiss={props.onDismiss} />
            ) : (
                <ManagePackageFiltersModalContent setActiveFilter={setActiveFilter} onDismiss={props.onDismiss} />
            )}
        </Modal>
    )
}
