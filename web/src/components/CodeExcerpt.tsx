import { highlightBlock } from 'highlight.js';
import * as React from 'react';
import * as VisibilitySensor from 'react-visibility-sensor';
import { fetchBlobContent } from 'sourcegraph/repo/backend';
import { getModeFromExtension, getPathExtension } from 'sourcegraph/util';
import { highlightNode } from 'sourcegraph/util/dom';
import { BlobPosition } from 'sourcegraph/util/types';

interface Props extends BlobPosition {
    // How many extra lines to show in the excerpt before/after the ref.
    previewWindowExtraLines?: number;
    highlightLength: number;
}

interface State {
    blobLines?: string[];
}

export class CodeExcerpt extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = {};
    }

    public getPreviewWindowLines(): number[] {
        const targetLine = this.props.line;
        let res = [targetLine];
        for (let i = targetLine - this.props.previewWindowExtraLines!; i < targetLine + this.props.previewWindowExtraLines! + 1; ++i) {
            if (i > 0 && i < targetLine) {
                res = [i].concat(res);
            }
            if (this.state.blobLines) {
                if (i < this.state.blobLines!.length && i > targetLine) {
                    res = res.concat([i]);
                }
            } else {
                if (i > targetLine) {
                    res = res.concat([i]);
                }
            }
        }
        return res;
    }

    public onChangeVisibility(isVisible: boolean): void {
        if (isVisible) {
            fetchBlobContent({repoPath: this.props.uri, commitID: this.props.rev, filePath: this.props.path}).then(content => {
                if (content) {
                    const blobLines = content.split('\n');
                    this.setState({ blobLines });
                }
            }).catch(e => {
                // TODO(slimsag): display error in UX
                console.error('failed to fetch blob content', e);
            });
        }
    }

    public render(): JSX.Element | null {
        return (
            <VisibilitySensor onChange={isVisible => this.onChangeVisibility(isVisible)} partialVisibility={true}>
                <table className='code-excerpt'>
                    <tbody>
                        {
                            this.getPreviewWindowLines().map(i =>
                                <tr key={i}>
                                    <td className='code-excerpt__line-number'>{i + 1}</td>
                                    <td className={'code-excerpt__code-line ' + getModeFromExtension(getPathExtension(this.props.path))}
                                        ref={!this.state.blobLines ? undefined : el => {
                                            if (el) {
                                                highlightBlock(el);
                                                if (i === this.props.line) {
                                                    highlightNode(el, this.props.char!, this.props.highlightLength);
                                                }
                                            }
                                        }}>
                                        {
                                            this.state.blobLines
                                            ? this.state.blobLines[i]
                                            : ' ' // create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch)
                                        }
                                    </td>
                                </tr>
                            )
                        }
                    </tbody>
                </table>
            </VisibilitySensor>
        );
    }
}
