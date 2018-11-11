import React from 'react'
import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from 'reactstrap'
import storage from '../../../browser/storage'
import { isSourcegraphDotCom } from '../../util/context'
import {
    LEARN_MORE_URL,
    QUICK_START_URL,
    SERVER_INSTALLATION_BODY,
    SERVER_INSTALLATION_HEADER,
    SERVER_INSTALLATION_LINK_CTA,
    SERVER_INSTALLATION_PRIMARY_CTA,
} from './utils'

interface State {
    modal: boolean
}

export class ServerModal extends React.Component<{}, State> {
    public state = {
        modal: false,
    }

    public componentDidMount(): void {
        storage.getSync(items => {
            this.setState(() => ({ modal: isSourcegraphDotCom(items.sourcegraphURL) && !items.hasSeenServerModal }))
        })
    }

    private toggle = () => {
        this.setState({
            modal: !this.state.modal,
        })
        storage.setSync({ hasSeenServerModal: true })
    }

    private openQuickStart = () => {
        window.open(QUICK_START_URL, '_blank')
        this.toggle()
    }

    private openLearnMore = () => {
        window.open(LEARN_MORE_URL, '_blank')
        this.toggle()
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <Modal isOpen={this.state.modal} toggle={this.toggle}>
                    <ModalHeader toggle={this.toggle}>{SERVER_INSTALLATION_HEADER}</ModalHeader>
                    <ModalBody>{SERVER_INSTALLATION_BODY}</ModalBody>
                    <ModalFooter className="options__footer">
                        <Button className="options__cta" color="primary" onClick={this.openQuickStart}>
                            {SERVER_INSTALLATION_PRIMARY_CTA}
                        </Button>{' '}
                        <Button className="options__link" color="link" onClick={this.openLearnMore}>
                            {SERVER_INSTALLATION_LINK_CTA}
                        </Button>
                    </ModalFooter>
                </Modal>
            </div>
        )
    }
}
