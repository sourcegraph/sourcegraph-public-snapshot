import React, { useMemo } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIcon } from '../icons'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { Tabs, TabList, Tab, TabPanels, TabPanel } from '@reach/tabs'
import { exampleCampaignSpecs } from './exampleCampaignSpecs'

const sourcePreviewCommand = 'src campaign preview -f hello-world.campaign.yaml -namespace sqs'

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
                <h2>1. Write a campaign spec YAML file</h2>
                <p>
                    The campaign spec (
                    <a
                        href="https://docs.sourcegraph.com/user/campaigns#campaign-specs"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </a>
                    ) describes what the campaign does. You'll provide it when previewing, creating, and updating
                    campaigns. We recommend committing it to source control.
                </p>
                <Tabs>
                    <TabList className="align-items-center">
                        <span className="font-weight-bold text-muted px-1">Examples:</span>
                        {exampleCampaignSpecs.map(({ name }) => (
                            <Tab key={name}>{name}</Tab>
                        ))}
                    </TabList>
                    <TabPanels>
                        {exampleCampaignSpecs.map(({ yaml }) => (
                            <TabPanel key={name}>
                                <div className="border p-2 mb-3">
                                    <pre className="m-0">{yaml}</pre>
                                </div>
                            </TabPanel>
                        ))}
                    </TabPanels>
                </Tabs>
                <h2 className="mt-4">2. Preview the campaign</h2>
                <p>
                    Use the{' '}
                    <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        Sourcegraph CLI (src)
                    </a>{' '}
                    to preview the commits and changesets that your campaign will make:
                </p>
                <pre>
                    <code>{sourcePreviewCommand}</code>
                </pre>
                <p>
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the
                    campaign.
                </p>
                <hr className="mt-5" />
                <p className="mt-2 text-muted">
                    Want more help? See{' '}
                    <a href="/help/user/campaigns" rel="noopener noreferrer" target="_blank">
                        campaigns documentation
                    </a>
                    .
                </p>
            </div>
        </>
    )
}
