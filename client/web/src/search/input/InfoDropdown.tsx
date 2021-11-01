import classNames from 'classnames'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

interface QueryFieldExample {
    /** A markdown string describing the example. */
    description: string
    /** The value for the example. Will be displayed as an inline code block. */
    value: string
}

import styles from './InfoDropdown.module.scss'

interface Props {
    title: string
    markdown: string
    examples?: QueryFieldExample[]
}

interface State {
    isOpen: boolean
}

export class InfoDropdown extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = { isOpen: false }
    }

    private toggleIsOpen = (): void => this.setState(previousState => ({ isOpen: !previousState.isOpen }))

    public render(): JSX.Element | null {
        return (
            <Dropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className={classNames('d-flex', styles.infoDropdown)}
            >
                <>
                    <DropdownToggle
                        tag="span"
                        caret={false}
                        className="pl-2 pr-0 btn btn-link d-flex align-items-center"
                    >
                        <HelpCircleOutlineIcon className="icon-inline small" />
                    </DropdownToggle>
                    <DropdownMenu right={true} className={classNames('pb-0', styles.item)}>
                        <DropdownItem header={true}>
                            <strong>{this.props.title}</strong>
                        </DropdownItem>
                        <DropdownItem divider={true} />
                        <div className={styles.content}>
                            <small dangerouslySetInnerHTML={{ __html: renderMarkdown(this.props.markdown) }} />
                        </div>

                        {this.props.examples && (
                            <>
                                <DropdownItem divider={true} />
                                <DropdownItem header={true}>
                                    <strong>{pluralize('Example', this.props.examples.length)}</strong>
                                </DropdownItem>
                                <ul className="list-unstyled mb-2">
                                    {this.props.examples.map(example => (
                                        <div key={example.value}>
                                            <div className="p-2">
                                                <span className="text-muted small">{example.description}: </span>
                                                <code>{example.value}</code>
                                            </div>
                                        </div>
                                    ))}
                                </ul>
                            </>
                        )}
                    </DropdownMenu>
                </>
            </Dropdown>
        )
    }
}
