import { useState } from 'react'

import { Modal, PageHeader } from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import { SiteAdminPackageFields } from '../../graphql-operations'

import { MultiPackageForm } from './MultiPackageForm'
import { SinglePackageForm } from './SinglePackageForm'

import styles from './BlockPackageModal.module.scss'

interface BlockPackagesModalProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    onDismiss: () => void
}

export type BlockType = 'single' | 'multiple'
export type Behaviour = 'blocklist' | 'allowlist'

export const BlockPackagesModal: React.FunctionComponent<BlockPackagesModalProps> = ({ node, filters, onDismiss }) => {
    const [blockType, setBlockType] = useState<BlockType>('single')
    const [behaviour, setBehaviour] = useState<Behaviour>('blocklist')

    return (
        <Modal aria-label="Manage package versions" onDismiss={onDismiss} className={styles.modal}>
            <PageHeader
                path={[{ text: 'Manage packages' }]}
                headingElement="h2"
                {...(behaviour === 'blocklist'
                    ? {
                          byline: 'Blocklisting will remove all specified packages from this instance, and prevent them from being synced in future.',
                      }
                    : {
                          byline: 'Allowlisting will remove all unspecified packages from this instance, and prevent them from being synced in future. ',
                      })}
                className={styles.header}
                // actions={
                //     // <Button variant="link" size="sm">
                //     //     Switch to allowlist
                //     // </Button>
                // }
            />
            <div className={styles.content}>
                {blockType === 'single' ? (
                    <SinglePackageForm node={node} filters={filters} setType={setBlockType} onDismiss={onDismiss} />
                ) : (
                    <MultiPackageForm node={node} filters={filters} setType={setBlockType} onDismiss={onDismiss} />
                )}
            </div>
        </Modal>
    )
}
