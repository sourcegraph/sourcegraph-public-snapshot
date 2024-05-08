# Publishing a new release

1. Increment the `version` in [`package.json`](package.json).
2. Commit the version increment, e.g. `Cody context filters test dataset: Release 1.0.0`.
3. Log in to npm by running `npm login`. Use credentials stored
   in [1Password](https://start.1password.com/open/i?a=HEDEDSLHPBFGRBTKAKJWE23XX4&v=dnrhbauihkhjs5ag6vszsme45a&i=oye4u4faaxmxxesugzqxojr4q4&h=team-sourcegraph.1password.com).
4. Run `pnpm publish`
