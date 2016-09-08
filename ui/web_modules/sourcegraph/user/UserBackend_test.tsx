import expect from "expect.js";
import {AuthInfo, User} from "sourcegraph/api";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {ExternalToken} from "sourcegraph/user";
import * as UserActions from "sourcegraph/user/UserActions";
import {UserBackend} from "sourcegraph/user/UserBackend";
import {immediateSyncPromise} from "sourcegraph/util/testutil/immediateSyncPromise";

const sampleAuthInfo: AuthInfo = {UID: 1, Login: "u", Write: false, Admin: false};
const sampleToken: ExternalToken = {uid: 1, host: "example.com", scope: "s"};
const sampleUser: User = {UID: 1, Login: "u", Betas: [], BetaRegistered: false} as any;

describe("UserBackend", () => {
	describe("should handle WantAuthInfo", () => {
		it("with authInfo available, no included user", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 200, json: () => sampleAuthInfo});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([new UserActions.FetchedAuthInfo("t", sampleAuthInfo)]);
		});
		it("with authInfo available, with included GitHub token and user and emails", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 200, json: () => Object.assign({}, sampleAuthInfo, {
					GitHubToken: sampleToken,
					IncludedUser: sampleUser,
				})});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([
				new UserActions.FetchedUser(sampleUser.UID as number, sampleUser),
				new UserActions.FetchedAuthInfo("t", sampleAuthInfo),
			]);
		});
		it("with authInfo unexpected error", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/auth-info");
				return immediateSyncPromise({status: 500, text: () => immediateSyncPromise("error", true)});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([]);
		});
	});
});
