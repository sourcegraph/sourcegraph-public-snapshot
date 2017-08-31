import * as React from 'react';
import { Link } from 'react-router-dom';
import * as url from 'sourcegraph/util/url';

export interface Props {
    path: string;
    partToUrl: (i: number) => string;
    partToClassName?: (i: number) => string;
}

export class Breadcrumb extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        const parts = this.props.path.split('/');
        const spans: JSX.Element[] = [];
        for (const [i, part] of parts.entries()) {
            const link = this.props.partToUrl(i);
            const className = `part ${this.props.partToClassName ? this.props.partToClassName(i) : ''}`;
            spans.push(<Link key={i} className={className} to={link}>{part}</Link>);
            if (i < parts.length - 1) {
                spans.push(<span key={'sep' + i} className='separator'>/</span>);
            }
        }
        return <span className='breadcrumb'>
            {...spans}
        </span>;
    }
}

export interface RepoBreadcrumbProps {
    uri: string;
    rev?: string;
    path?: string;
}

export class RepoBreadcrumb extends React.Component<RepoBreadcrumbProps, {}> {
    public render(): JSX.Element | null {
        const trimmedUri = this.props.uri.split('/').slice(1).join('/'); // remove first path part
        return <Breadcrumb path={trimmedUri + (this.props.path ? '/' + this.props.path : '')} partToUrl={this.partToUrl} partToClassName={this.partToClassName} />;
    }

    private partToUrl = (i: number) => {
        const trimmedUri = this.props.uri.split('/').slice(1).join('/'); // remove first path part
        const uriParts = trimmedUri.split('/');
        if (i < uriParts.length) {
            return '/' + this.props.uri;
        }
        if (this.props.path) {
            const j = i - uriParts.length;
            const pathParts = this.props.path.split('/');
            if (j < pathParts.length - 1) {
                return url.toTree({ uri: this.props.uri, rev: this.props.rev, path: pathParts.slice(0, j + 1).join('/') });
            }
            return url.toBlob({ uri: this.props.uri, rev: this.props.rev, path: this.props.path });
        }
        return '';
    }

    private partToClassName = (i: number) => {
        const trimmedUri = this.props.uri.split('/').slice(1).join('/'); // remove first path part
        const uriParts = trimmedUri.split('/');
        if (i < uriParts.length) {
            return 'part-repo';
        }
        if (this.props.path) {
            const j = i - uriParts.length;
            const pathParts = this.props.path.split('/');
            if (j < pathParts.length - 1) {
                return 'part-directory';
            }
            return 'part-file';
        }
        return '';
    }
}
