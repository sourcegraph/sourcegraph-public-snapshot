import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DashboardBackend from "sourcegraph/dashboard/DashboardBackend";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import * as AlertActions from "sourcegraph/alerts/AlertActions";

describe("DashboardBackend", () => {
	it("should handle WantCreateRepo", () => {
		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.ui/.repo-create?RepoURI=aname");
			callback(null, {statusCode: 200}, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantCreateRepo("aname"));
		})).to.eql([new DashboardActions.RepoCreated("someFile")]);
	});
});

describe("DashboardBackend", () => {
	it("should handle WantddMirrorRepos", () => {
		let expectedURI = "/.ui/.repo-mirror";

		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			expect(options.json).to.have.property("Repos", "someRepos");
			callback(null, {statusCode: 200}, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantAddMirrorRepos("someRepos"));
		})).to.eql([new DashboardActions.MirrorReposAdded("someFile")]);
	});
});

describe("DashboardBackend", () => {
	it("should handle WantAddMirrorRepo", () => {
		let expectedURI = "/.ui/.repo-mirror";

		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			expect(options.json).to.have.property("Repos");
			callback(null, {statusCode: 200}, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantAddMirrorRepo("someRepos"));
		})).to.eql([new DashboardActions.MirrorRepoAdded("someRepos", "someFile")]);
	});
});

describe("DashboardBackend", () => {
	it("should handle WantInviteUser", () => {
		let action = {
			email: "123@abc.com",
			permission: "admin",
		};
		let expectedURI = `/.ui/.invite`;

		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			expect(options.json).to.have.property("Email");
			expect(options.json).to.have.property("Permission");
			callback(null, {statusCode: 200}, {Link: "hello"});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantInviteUser(action.email, action.permission));
		})).to.eql([new DashboardActions.UserInvited({
			Name: action.email,
			Admin: true,
			Write: false,
		}), new AlertActions.AddAlert(
			false,
			`Please send <a href="hello">this invitation link</a> to <strong>123@abc.com</strong>.`
		)]);
	});
});

describe("DashboardBackend", () => {
	it("should handle WantInviteUsers", () => {
		let expectedURI = "/.ui/.invite-bulk";

		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			expect(options.json).to.have.property("Emails");
			callback(null, {statusCode: 200}, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantInviteUsers("someEmails"));
		})).to.eql([new DashboardActions.UsersInvited("someFile")]);
	});
});
