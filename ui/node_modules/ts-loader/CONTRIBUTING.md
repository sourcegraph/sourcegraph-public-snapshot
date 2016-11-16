# Contributer's Guide

We welcome contributions from the community and have gathered guidelines 
here to help you get started.

## Discussion

While not absolutely required, it is encouraged that you first open an issue 
for any bug or feature request. This allows discussion on the proper course of
action to take before coding begins.

## Building

```shell
npm install
npm run build
```

## Changing

Most of the information you need to contribute code changes can [be found here](https://guides.github.com/activities/contributing-to-open-source/).
In short: fork, make your changes, and submit a pull request.

## Testing

This project makes use of 2 integration test packs to make sure we don't break anything. That's right, count them, 2! There is a comparison test pack which compares compilation outputs and is long established.  There is also an execution test pack which executes the compiled JavaScript. This test pack is young and contains fewer tests; but it shows promise.

You can run all the tests (in both test packs) with `npm test`.

To run comparison tests alone use `npm run comparison-tests`.
To run execution tests alone use `npm run execution-tests`.

Not all bugs/features necessarily fit into either framework and that's OK. However, most do and therefore you should make every effort to create at least one test which demonstrates the issue or exercises the feature. Use your judgement to decide whether you think a comparison test or an execution test is most appropriate.

To read about the comparison test pack take a look [here](test/comparison-tests/README.md)
To read about the execution test pack take a look [here](test/execution-tests/README.md)

## Publishing

So the time has come to publish the latest version of ts-loader to npm. Exciting!

Before you can actually publish make sure the following statements are true:

- Tests should be green
- The version number in [package.json](package.json) has been incremented.
- The [changelog](CHANGELOG.md) has been updated with details of the changes in this release.  Where possible include the details of the issues affected and the PRs raised.

OK - you're actually ready.  We're going to publish.  Here we need tread carefully. Follow these steps: 

- clone ts-loader from the main repo with this command: `git clone https://github.com/TypeStrong/ts-loader.git`
- [Login to npm](https://docs.npmjs.com/cli/adduser) if you need to: `npm login`
- install ts-loaders packages with `npm install`
- build ts-loader with `npm run build`
- run the tests to ensure all is still good: `npm test`

If all the tests passed then we're going to ship:
- tag the release in git.  You can see existing tags with the command `git tag`.  If the version in your `package.json` is `"1.0.1"` then you would tag the release like so: `git tag v1.0.1`.  For more on type of tags we're using read [here](https://git-scm.com/book/en/v2/Git-Basics-Tagging#Lightweight-Tags).
- Push the tag so the new version will show up in the [releases](https://github.com/TypeStrong/ts-loader/releases): `git push origin --tags`
- On the releases page, click the "Draft a new release button" and, on the presented page, select the version you've just released, name it and copy in the new markdown that you added to the [changelog](CHANGELOG.md).
- Now the big moment: `npm publish`

You've released!  Pat yourself on the back.