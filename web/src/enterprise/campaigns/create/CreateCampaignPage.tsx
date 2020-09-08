import React, { useMemo } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIcon } from '../icons'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'

const campaignSpec = `name: hello-world
description: Add Hello World to READMEs

# Find all repositories that contain a README.md file.
on:
  - repositoriesMatchingQuery: file:README.md

# In each repository, run this command. Each repository's resulting diff is captured.
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Hello World
  body: My first campaign!
  branch: hello-world # Push the commit to this branch.
  commit:
    message: Append Hello World to all README.md files
  published: false`

const helloWorldDownloadUrl = 'data:text/plain;charset=utf-8,' + encodeURIComponent(campaignSpec)

const sourcePreviewCommand =
    'src campaign preview -f hello-world.campaign.yaml -namespace sourcegraph-username-or-organisation'

export interface CreateCampaignPageProps extends BreadcrumbSetters {
    // Nothing for now, but using it so once this changes we get type errors in the routing files.
}

export const CreateCampaignPage: React.FunctionComponent<CreateCampaignPageProps> = ({ useBreadcrumb }) => {
    useBreadcrumb(useMemo(() => ({ element: <>Create campaign</>, key: 'createCampaignPage' }), []))
    return (
        <>
            <PageTitle title="Create campaign" />
            <PageHeader
                icon={CampaignsIcon}
                title={
                    <>
                        Create campaign{' '}
                        <sup>
                            <span className="badge badge-merged text-uppercase">Beta</span>
                        </sup>
                    </>
                }
            />
            <div className="container pt-3">
                <h2>New to campaigns?</h2>
                <p className="lead">
                    Read the{' '}
                    <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener noreferrer" target="_blank">
                        campaigns documentation page
                    </a>{' '}
                    to learn how to create campaign specifications, using Sourcegraph's CLI tool src-cli and publishing
                    changesets.
                </p>
                <h2>Quick start</h2>
                <div className="d-flex justify-content-between align-items-center mb-2">
                    <p className="m-0 lead">This campaign specification adds "Hello World" to all README.md files:</p>
                    <a
                        download="hello-world.campaign.yaml"
                        href={helloWorldDownloadUrl}
                        className="text-right btn btn-secondary text-nowrap"
                        data-tooltip="Download hello-world.campaign.yaml"
                    >
                        <FileDownloadIcon className="icon-inline" /> Download yaml
                    </a>
                </div>
                <div className="bg-light rounded p-2 mb-3">
                    <pre className="m-0">{campaignSpec}</pre>
                </div>
                <p className="lead">
                    Use Sourcegraph's CLI tool, <code>src</code>, to execute the steps in the campaign spec and upload
                    it, ready to be previewed and applied:
                </p>
                <div className="bg-light rounded p-3 mb-3">
                    <pre className="m-0">{sourcePreviewCommand}</pre>
                </div>
                <p className="lead">
                    Download <code>src</code> at{' '}
                    <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        github.com/sourcegraph/src-cli
                    </a>
                </p>
            </div>
        </>
    )
}
