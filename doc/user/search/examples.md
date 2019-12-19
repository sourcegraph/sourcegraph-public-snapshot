# Search examples

Below are examples that search repositories on [Sourcegraph.com](https://sourcegraph.com/search), our open source code search solution for GitHub and GitLab. You can copy and adapt the following search queries for use on your company’s private instance.

> See [**search query syntax**](queries.md) reference.

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
