import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import IssueBackend from "sourcegraph/issue/IssueBackend";
import * as IssueActions from "sourcegraph/issue/IssueActions";

describe("IssueBackend", () => {
	it("should handle IssueCreate", () => {
		IssueBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/aRepo/.tracker/new");
			callback(null, null, "sourcegraph.com/.tracker/42");
		};
		let cid = "a".repeat(40);
		Dispatcher.directDispatch(
			IssueBackend, new IssueActions.CreateIssue("aRepo", "a/path", cid, 1, 42, "aTitle", "aBody", function() {})
		);
	});
});
