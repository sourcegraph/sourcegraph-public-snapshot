// @flow

import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import UserBackend from "sourcegraph/user/UserBackend";
import * as UserActions from "sourcegraph/user/UserActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";
import type {AuthInfo, User} from "sourcegraph/user";

const sampleAuthInfo: AuthInfo = {UID: 1, Login: "u"};
const sampleUser: User = {UID: 1, Login: "u"};

describe("UserBackend", () => {
	describe("should handle WantAuthInfo", () => {
		it("with authInfo available, no included user", () => {
			UserBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 200, json: () => sampleAuthInfo});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([new UserActions.FetchedAuthInfo("t", sampleAuthInfo)]);
		});
		it("with authInfo available, with included user", () => {
			UserBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 200, json: () => ({...sampleAuthInfo, IncludedUser: sampleUser})});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([
				new UserActions.FetchedAuthInfo("t", sampleAuthInfo),
				new UserActions.FetchedUser(sampleUser.UID, sampleUser),
			]);
		});
		it("with authInfo unexpected error", () => {
			UserBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 500, text: () => immediateSyncPromise("error", true)});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([new UserActions.FetchedAuthInfo("t", {Error: "error"})]);
		});
	});
	describe("should handle WantUser", () => {
		it("with user available", () => {
			UserBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/users/1$");
				return immediateSyncPromise({status: 200, json: () => sampleUser});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantUser(1));
			})).to.eql([new UserActions.FetchedUser(1, sampleUser)]);
		});
		it("with user not available", () => {
			UserBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/users/1$");
				return immediateSyncPromise({status: 404, text: () => immediateSyncPromise("error", true)});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantUser(1));
			})).to.eql([new UserActions.FetchedUser(1, {Error: "error"})]);
		});
	});
});
