import { highlightBlock } from 'highlight.js';
import * as React from 'react';
import * as VisibilitySensor from 'react-visibility-sensor';
import { fetchBlobContent } from 'sourcegraph/backend';
import { getModeFromExtension, getPathExtension } from 'sourcegraph/util';
import * as colors from 'sourcegraph/util/colors';
import { highlightNode } from 'sourcegraph/util/dom';
import { BlobPosition } from 'sourcegraph/util/types';
import { classes, style } from 'typestyle';

interface Props extends BlobPosition {
    // How many extra lines to show in the excerpt before/after the ref.
    previewWindowExtraLines?: number;
    highlightLength: number;
}

interface State {
    blobLines?: string[];
}

namespace Styles {
    export const lineNum = style({ color: colors.baseColor, paddingRight: '15px' });
    export const codeLine = style({ whiteSpace: 'pre' });
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
            fetchBlobContent(this.props.uri, this.props.rev, this.props.path).then(content => {
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
        return <VisibilitySensor onChange={isVisible => this.onChangeVisibility(isVisible)} partialVisibility={true}>
            <table>
                <tbody>
                    {
                        this.getPreviewWindowLines().map(i =>
                            <tr key={i}>
                                <td className={Styles.lineNum}>{i + 1}</td>
                                <td className={classes(getModeFromExtension(getPathExtension(this.props.path)), Styles.codeLine)}
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
        </VisibilitySensor>;
    }
}
