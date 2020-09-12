export const exampleCampaignSpecs: {
    name: string
    yaml: string
}[] = [
    {
        name: 'hello-world.campaign.yaml',
        yaml: `name: hello-world
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
  published: false`,
    },
    {
        name: 'minimal.campaign.yaml',
        yaml: `name: my-campaign

# Add your own description, query, steps, changesetTemplate, etc.
# See https://docs.sourcegraph.com/user/campaigns for help.`,
    },
]
