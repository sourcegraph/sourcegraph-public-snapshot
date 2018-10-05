import { ActionItem } from '@sourcegraph/extensions-client-common/lib/app/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/extensions-client-common/lib/app/actions/ActionsContainer'
import BookOpenIcon from 'mdi-react/BookOpenIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import LinkIcon from 'mdi-react/LinkIcon'
import SquareEditOutlineIcon from 'mdi-react/SquareEditOutlineIcon'
import StarIcon from 'mdi-react/StarIcon'
import * as React from 'react'
import { fromEvent, merge, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { ContributableMenu } from 'sourcegraph/module/protocol'
import { Key } from 'ts-key-enum'
import { ExtensionsControllerProps, ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'

interface Props extends ExtensionsControllerProps, ExtensionsProps {
    onDismiss: () => void
}

export class HelpPopover extends React.Component<Props> {
    private static LINKS: { title: string; description: string; template: string }[] = [
        { title: 'Bug report', description: 'Report problems and unexpected behavior', template: 'bug_report.md' },
        { title: 'Feature request', description: 'Suggest an idea for Sourcegraph', template: 'feature_request.md' },
        { title: 'Question', description: 'Ask a question about Sourcegraph', template: 'question.md' },
    ]

    private subscriptions = new Subscription()

    private ref: HTMLElement | null = null
    private setRef = (e: HTMLElement | null) => (this.ref = e)

    public componentDidMount(): void {
        const escKeypress = fromEvent<KeyboardEvent>(window, 'keydown').pipe(filter(event => event.key === Key.Escape))

        const outsideClick = fromEvent<MouseEvent>(window, 'mousedown').pipe(
            filter(event => !!this.ref && !this.ref.contains(event.target as HTMLElement))
        )

        this.subscriptions.add(merge(escKeypress, outsideClick).subscribe(() => this.props.onDismiss()))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        return (
            <div className="help-popover card" ref={this.setRef}>
                <h4 className="card-header d-flex justify-content-between pl-3">
                    Help
                    <button
                        type="reset"
                        className="btn btn-icon help-popover__close"
                        onClick={this.props.onDismiss}
                        title="Close"
                    >
                        <CloseIcon className="icon-inline" />
                    </button>
                </h4>
                <div className="list-group list-group-flush">
                    <a
                        className="list-group-item list-group-item-action px-3 py-2"
                        href="https://about.sourcegraph.com/docs/search/query-syntax"
                        target="_blank"
                    >
                        <StarIcon className="icon-inline" /> How to search
                    </a>
                    <a
                        className="list-group-item list-group-item-action px-3 py-2"
                        href="https://about.sourcegraph.com/docs"
                        target="_blank"
                    >
                        <BookOpenIcon className="icon-inline" /> Sourcegraph documentation
                    </a>
                    <a
                        className="list-group-item list-group-item-action px-3 py-2"
                        href="https://github.com/sourcegraph/issues/issues"
                        target="_blank"
                        rel="noreferrer"
                    >
                        <LinkIcon className="icon-inline" /> Public issue tracker
                    </a>
                    <a
                        className="list-group-item list-group-item-action px-3 py-2"
                        href="https://about.sourcegraph.com/contact"
                        target="_blank"
                    >
                        <SquareEditOutlineIcon className="icon-inline" /> Contact Sourcegraph
                    </a>
                </div>
                <h4 className="card-header pl-3">File a public issue...</h4>
                <div className="list-group list-group-flush">
                    {HelpPopover.LINKS.map(({ title, description, template }, i) => (
                        <a
                            className="list-group-item list-group-item-action px-3 py-2"
                            href={`https://github.com/sourcegraph/issues/issues/new?template=${template}`}
                            target="_blank"
                            rel="noreferrer"
                        >
                            <strong>{title} &raquo;</strong>
                            <br />
                            <small className="text-muted">{description}</small>
                        </a>
                    ))}
                </div>
                <ActionsContainer
                    menu={ContributableMenu.Help}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={items => (
                        <>
                            <h4 className="card-header pl-3">Extensions help</h4>
                            {items.map((item, i) => (
                                <ActionItem
                                    key={i}
                                    {...item}
                                    extensionsController={this.props.extensionsController}
                                    extensions={this.props.extensions}
                                />
                            ))}
                        </>
                    )}
                    empty={null}
                    extensionsController={this.props.extensionsController}
                    extensions={this.props.extensions}
                />
            </div>
        )
    }
}
