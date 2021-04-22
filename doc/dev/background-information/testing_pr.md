# Testing a pull request 

When submitting a pull request to a Sourcegraph-maintained extension repository, there are a few tests you'll need to pass before the request can be reviewed. These tests run logical tests that we author to make sure our code runs as expected.

## What are the tests?
- **Prettier** is responsible for mostly visual choices like if we have a line of code that's longer than 70 characters, it will fail. This has nothing to do with the correctness of the code, more of a stylistic choice we make as a team to make sure we all use the same standard.
- **eslint** is similar but more so focused on the logic aspect of the code. We use eslint to avoid bad patterns which otherwise stylistically make sense but lead to potentially buggy code
- **Husky** ([link](https://typicode.github.io/husky/#/)) and **Semantic Pull Requests** ([link](https://github.com/zeke/semantic-pull-requests)) ensure your commit messages have enough semantic information to be able to trigger a release. For example, `feat:` tag is for new features.
- Update README.md and manifest accordingly

## Run the following commands before submitting your Pull Request:
1. `npm run prettier` : applies prettier to your code
1. `npm run prettier-check`: checks if your code is passing the prettier checks
1. `npm run eslint`: checks if your code is passing the eslint checks


## After the Pull Request has been submitted:
1. Add `team/extensibility` to label
2. Edit the PR title by adding a semantic prefix like `fix:` or `feat:` to the front. See more examples on our [GitHub Pull Requests](https://github.com/sourcegraph/sourcegraph-vscode/pulls) page.
