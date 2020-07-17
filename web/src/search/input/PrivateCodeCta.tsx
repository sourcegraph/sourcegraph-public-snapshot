import * as React from 'react'
import DownloadIcon from 'mdi-react/DownloadIcon'

export const PrivateCodeCta: React.FunctionComponent<{}> = () => {
    const logoSource = '/.assets/img/sourcegraph-mark.svg'
    return (
        <div className="private-code-cta d-flex flex-row card">
            <div className="private-code-cta__icon-column d-flex flex-column align-items-center">
                <img src={logoSource} className="private-code-cta__logo" />
                <DownloadIcon className="private-code-cta__download-icon" />
            </div>
            <div>
                <h3>Search your private code</h3>
                <p>
                    Set up a private Sourcegraph instance to search your private repositories on GitHub, GitLab,
                    Bitbucket and local installations of git, perforce, svn and other code repositories.
                </p>
                <div>
                    <button className="btn btn-primary">Install now</button>
                </div>
            </div>
        </div>
    )
}
