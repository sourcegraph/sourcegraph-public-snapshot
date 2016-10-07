import expect from "expect.js";
import { URIUtils } from "sourcegraph/core/uri";
import URI from "vs/base/common/uri";

describe("URIUtils", () => {
	const tests = [
		{repo: "github.com/a/b", rev: "v", path: "p", str: "git://github.com/a/b?v#p"},
		{repo: "github.com/a/b", rev: "", path: "p", str: "git://github.com/a/b?HEAD#p"},
		{repo: "github.com/a/b", rev: "v", path: "", str: "git://github.com/a/b?v"},
		{repo: "github.com/a/b", rev: "", path: "", str: "git://github.com/a/b?HEAD"},
		{repo: "github.com/a/b", rev: null, path: "", str: "git://github.com/a/b?HEAD"},
	];
	describe("pathInRepo", () => {
		tests.forEach(test => {
			it(JSON.stringify(test), () => {
				expect(URIUtils.pathInRepo(test.repo, test.rev, test.path).toString()).to.eql(test.str);
			});
		});
	});
	describe("repoParams", () => {
		tests.forEach(test => {
			it(JSON.stringify(test), () => {
				expect(URIUtils.repoParams(URI.parse(test.str))).to.eql({repo: test.repo, rev: test.rev || "HEAD", path: test.path});
			});
		});
	});
});
