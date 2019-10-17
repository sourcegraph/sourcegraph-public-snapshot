# Saved searches

Saved searches lets you save and describe search queries so you can easily monitor the results on an ongoing basis. You can create a saved search for anything, including diffs and commits across all branches of your repositories.

Saved searches can be an early warning system for common problems in your code--and a way to monitor best practices, the progress of refactors, etc. Alerts for saved searches can be sent through email, ensuring you're aware of important code changes.

## Example saved searches

- [Recent security-related changes on all branches](https://sourcegraph.com/search?q=type:diff+repo:github%5C.com/kubernetes/kubernetes%24+repo:%40*refs/heads/+after:"5+days+ago"+%5Cb%28auth%5B%5Eo%5D%5B%5Er%5D%7Csecurity%5Cb%7Ccve%7Cpassword%7Csecure%7Cunsafe%7Cperms%7Cpermissions%29)<br/>
`type:diff repo:@*refs/heads/ after:"5 days ago" \b(auth[^o][^r]|security\b|cve|password|secure|unsafe|perms|permissions)`

- [Admitted hacks and TODOs in app code](https://sourcegraph.com/search?q=repogroup:sample+-file:%5C.%28json%7Cmd%7Ctxt%29%24+hack%7Ctodo%7Ckludge%7Cfixme)<br/>
`-file:\.(json|md|txt)$ hack|todo|kludge|fixme`

- [New usages of a function](https://sourcegraph.com/search?q=repo:github%5C.com/sourcegraph/+type:diff+after:%221+week+ago%22+%5C.subscribe%5C%28+lang:typescript)<br/>
`type:diff after:"1 week ago" \.subscribe\( lang:typescript`

- [Recent quality related changes on all branches (customize for your linters)](https://sourcegraph.com/search?q=repo:github%5C.com/sourcegraph/+repo:%40*refs/heads/:%5Emaster+type:diff+after:"1+week+ago"+%28tslint:disable%29)<br/>
`repo:@*refs/heads/:^master type:diff after:"1 week ago" (tslint:disable)`

- [Recent dependency changes](https://sourcegraph.com/search?q=repo:github%5C.com/sourcegraph/+file:package.json+type:diff+after:%221+week+ago%22)<br/>
`file:package.json type:diff after:"1 week ago"`

## Creating saved searches

A saved search consists of a description and a query, both of which you can define and edit.

Saved searches are written as JSON entries in settings, and they can be associated with a user or an org:

- User saved searches are only visible to (and editable by) the user that created them.
- Org saved searches are visible to (and editable by) all members of the org.

To create a saved search:

1. Go to **User menu > Saved searches** in the top navigation bar.
1. Press the **+ Add new search** button.
1. In the **Query** field, type in the components of the search query.
1. In the **Description** field, type in a human-readable description for your saved search.
1. In the user and org dropdown menu, select where you'd like the search to be saved.
1. Click **Create**. The saved search is created, and you can see the number of results.

Alternatively, to create a saved search from a search you've already run:

1. Execute a search from the homepage or navigation bar.
1. Press the **Save this search query** button that appears on the right side of the screen above the first result.
1. Follow the instructions from above to fill in the remaining fields.

To view saved searches, go to **User menu > Saved searches** in the top navigation bar.

## Configuring email notifications

Sourcegraph can automatically run your saved searches and notify you when new results are available via email. With this feature you can get notified about issues in your code (such as licensing issues, security changes, potential secrets being committed, etc.)

To configure email notifications, click **Edit** on a saved search and check the **Email notifications** checkbox and press **Save**. You will receive a notification telling you it is set up and working almost instantly!

By default, email notifications notify the owner of the configuration (either a single user or the entire org).
