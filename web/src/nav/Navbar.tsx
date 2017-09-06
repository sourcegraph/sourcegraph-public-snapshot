
import * as React from 'react';
import { Link } from 'react-router-dom';
import { SearchBox } from 'sourcegraph/search/SearchBox';
import { ParsedRouteProps } from 'sourcegraph/util/routes';
import { toEditorURL } from 'sourcegraph/util/url';

export class Navbar extends React.Component<ParsedRouteProps, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='navbar'>
                <Link to='/' className='navbar__logo-link'>
                    <img className='navbar__logo' src='/.assets/img/sourcegraph-mark.svg' />
                </Link>
                <div className='navbar__search-box-container'>
                    <SearchBox history={this.props.history} />
                </div>
                <div>
                    {
                        this.props.repoPath &&
                            <a href={toEditorURL(this.props.repoPath, this.props.commitID, this.props.filePath)} target='_blank' className='open-on-desktop'>
                                <span>Open on desktop</span>
                                <svg className='icon' width='11px' height='9px'>
                                    {/* tslint:disable-next-line:max-line-length */}
                                    <path fill='#FFFFFF' xmlns='http://www.w3.org/2000/svg' id='path10_fill' d='M 6.325 8.4C 6.125 8.575 5.8 8.55 5.625 8.325C 5.55 8.25 5.5 8.125 5.5 8L 5.5 6C 2.95 6 1.4 6.875 0.825 8.7C 0.775 8.875 0.6 9 0.425 9C 0.2 9 -4.44089e-16 8.8 -4.44089e-16 8.575C -4.44089e-16 8.575 -4.44089e-16 8.575 -4.44089e-16 8.55C 0.125 4.825 1.925 2.675 5.5 2.5L 5.5 0.5C 5.5 0.225 5.725 8.88178e-16 6 8.88178e-16C 6.125 8.88178e-16 6.225 0.05 6.325 0.125L 10.825 3.875C 11.025 4.05 11.075 4.375 10.9 4.575C 10.875 4.6 10.85 4.625 10.825 4.65L 6.325 8.4Z' />
                                </svg>
                            </a>
                    }
                </div>
            </div>
        );
    }
}
