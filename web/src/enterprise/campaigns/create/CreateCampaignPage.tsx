import React from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { CampaignHeader } from '../detail/CampaignHeader'
import MagnifyIcon from 'mdi-react/MagnifyIcon'

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

const sourceApplyCommand = 'src campaign apply -f hello-world.campaign.yaml -preview'

export interface CreateCampaignPageProps {
    // Nothing for now, but using it so once this changes we get type errors in the routing files.
}

export const CreateCampaignPage: React.FunctionComponent<CreateCampaignPageProps> = () => (
    <>
        <PageTitle title="Create campaign" />
        <CampaignHeader name="Create campaign" />
        <div className="container pt-3">
            <h2>New to campaigns?</h2>
            <p className="lead">
                Read the{' '}
                <a href="https://docs.sourcegraph.com" rel="noopener noreferrer" target="_blank">
                    creating a campaign
                </a>{' '}
                documentation page to learn about creating campaign specifications, using src-cli and publishing
                changesets.
            </p>
            <h2>Quick start</h2>
            <div className="d-flex justify-content-between align-items-center mb-2">
                <p className="m-0 lead">This campaign specification adds "Hello World" to all README.md files:</p>
                <a
                    download="hello-world.campaign.yaml"
                    href={helloWorldDownloadUrl}
                    className="text-right btn btn-secondary text-nowrap"
                >
                    <MagnifyIcon className="icon-inline" /> Download yaml
                </a>
            </div>
            <div className="bg-light rounded p-2 mb-3">
                <pre className="m-0">{campaignSpec}</pre>
            </div>
            <p className="lead">Use Sourcegraph src-cli to apply the campaign in preview mode:</p>
            <div className="bg-light rounded p-3 mb-3">
                <pre className="m-0">{sourceApplyCommand}</pre>
            </div>
            <p className="lead">
                Download Sourcegraph's cli tool, src-cli at{' '}
                <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                    github.com/sourcegraph/src-cli
                </a>
            </p>
        </div>
    </>
)
