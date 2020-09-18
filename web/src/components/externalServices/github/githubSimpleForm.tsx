import React from 'react'
import { SimpleExternalServiceForm } from '../externalServices'

interface Props {}

export const GitHubExternalServiceSimpleForm: React.FunctionComponent<Props> = () => (
    <>
        <div className="form-group">
            <label>GitHub personal access token</label>
            <input type="text" className="form-control text-monospace" size={40} />
            <small className="form-text text-muted">
                <a href="https://github.com/settings/tokens/new?description=Sourcegraph">
                    Create a new access token on GitHub.com
                </a>{' '}
                with the <strong>repo</strong> or <strong>public_repo</strong> scope.
            </small>
        </div>
        <div className="form-group">
            <label>Repositories to sync</label>
            <textarea className="form-control" rows={3} />
            <small className="form-text text-muted">
                Specify repositories as space-separated (or newline-separated) entries of the form{' '}
                <strong>owner/repo</strong>.<br />
                Example: <code>alice/my-project mycompany/my-excellent-repo</code>
            </small>
        </div>
        <div className="form-group">
            <label>Users or organizations to sync all repositories from</label>
            <textarea className="form-control" rows={3} />
            <small className="form-text text-muted">
                Specify as space-separated (or newline-separated) names of GitHub users or organizations.
                <br />
                Example: <code>alice mycompany</code>
            </small>
        </div>
    </>
)

export const githubSimpleForm: SimpleExternalServiceForm = {
    supportsConfig: (config: string) => true, // TODO(sqs)
    component: GitHubExternalServiceSimpleForm,
}
