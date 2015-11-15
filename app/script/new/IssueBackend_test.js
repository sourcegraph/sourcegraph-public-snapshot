import expect from "expect.js";

import Dispatcher from "./Dispatcher";
import IssueBackend from "./IssueBackend";
import * as IssueActions from "./IssueActions";

describe("IssueBackend", () => {
	it("should handle IssueCreate", () => {
		IssueBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/aRepo/.threads/new");
			callback(null, null, "sourcegraph.com/.threads/42");
		};
		let cid = "a".repeat(40);
		Dispatcher.directDispatch(
			IssueBackend, new IssueActions.CreateIssue("aRepo", "a/path", cid, 1, 42, "aTitle", "aBody", () => {})
		);
	});
});
