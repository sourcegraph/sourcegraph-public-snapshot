import { Modal } from '@sourcegraph/wildcard'
import type { ModalProps } from '@sourcegraph/wildcard/out/src/components/Modal'

type ShareLinkModalProps = ModalProps

export const ShareLinkModal: React.FunctionComponent<ShareLinkModalProps> = ({ isOpen, onDismiss }) => (
    <Modal aria-label="Share insight" isOpen={isOpen} onDismiss={onDismiss}>
        Share me!
    </Modal>
)
