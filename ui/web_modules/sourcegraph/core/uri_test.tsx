/*
 *import expect from "expect.js";
 *import { uriutils } from "sourcegraph/core/uri";
 *import uri from "vs/base/common/uri";
 *
 *describe("uriutils", () => {
 *  const tests = [
 *    {repo: "github.com/a/b", rev: "v", path: "p", str: "git://github.com/a/b?v#p"},
 *    {repo: "github.com/a/b", rev: "", path: "p", str: "git://github.com/a/b?#p"},
 *    {repo: "github.com/a/b", rev: "v", path: "", str: "git://github.com/a/b?v"},
 *    {repo: "github.com/a/b", rev: "", path: "", str: "git://github.com/a/b?"},
 *    {repo: "github.com/a/b", rev: null, path: "", str: "git://github.com/a/b?"},
 *  ];
 *  describe("pathinrepo", () => {
 *    tests.foreach(test => {
 *      it(json.stringify(test), () => {
 *        expect(uriutils.pathinrepo(test.repo, test.rev, test.path).tostring()).to.eql(test.str);
 *      });
 *    });
 *  });
 *  describe("repoparams", () => {
 *  tests.foreach(test => {
 *    it(json.stringify(test), () => {
 *      expect(uriutils.repoparams(uri.parse(test.str))).to.eql({repo: test.repo, rev: test.rev, path: test.path});
 *      });
 *    });
 *  });
 *});
 */
