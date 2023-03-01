import { useState } from 'react'

import { Button, H2, Modal } from '@sourcegraph/wildcard'

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

const MODAL_LABEL_ID = 'site-admin-packages-block-modal'

export type BlockType = 'single' | 'multiple'

export const BlockPackagesModal: React.FunctionComponent<BlockPackagesModalProps> = ({ node, filters, onDismiss }) => {
    const [blockType, setBlockType] = useState<BlockType>('single')

    return (
        <Modal aria-labelledby={MODAL_LABEL_ID} onDismiss={onDismiss} className={styles.modal}>
            <div className={styles.header}>
                <H2 id={MODAL_LABEL_ID}>Block package</H2>
            </div>
            <div className={styles.content}>
                {blockType === 'single' ? (
                    <SinglePackageForm node={node} filters={filters} setType={setBlockType} />
                ) : (
                    <MultiPackageForm node={node} filters={filters} setType={setBlockType} />
                )}
            </div>
            <div className={styles.actions}>
                <Button variant="secondary" onClick={onDismiss} className="mr-2">
                    Cancel
                </Button>
                <Button variant="danger" onClick={onDismiss}>
                    Block {blockType === 'single' ? 'package' : 'packages'}
                </Button>
            </div>
        </Modal>
    )
}
