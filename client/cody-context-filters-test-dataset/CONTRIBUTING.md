# Publishing a new release

1. Ensure the updated version of the test dataset is already committed to `main`.
2. Create a new branch from `main` and switch to it.
3. Bump the version in [`package.json`](package.json).
4. Commit the version change.
5. Run `pnpm login` and enter the credentials
   from [1Password](https://start.1password.com/open/i?a=HEDEDSLHPBFGRBTKAKJWE23XX4&v=dnrhbauihkhjs5ag6vszsme45a&i=oye4u4faaxmxxesugzqxojr4q4&h=team-sourcegraph.1password.com).
6. Run `pnpm publish` from the package root and check for any errors.
7. Once the release has been published, create a PR and get it merged to `main`.
