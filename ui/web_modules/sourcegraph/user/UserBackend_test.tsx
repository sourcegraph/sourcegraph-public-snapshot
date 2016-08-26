import expect from "expect.js";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {AuthInfo, EmailAddr, ExternalToken, User} from "sourcegraph/user/index";
import * as UserActions from "sourcegraph/user/UserActions";
import {UserBackend} from "sourcegraph/user/UserBackend";
import {immediateSyncPromise} from "sourcegraph/util/immediateSyncPromise";

const sampleAuthInfo: AuthInfo = {UID: 1, Login: "u", Write: false, Admin: false, IntercomHash: ""};
const sampleToken: ExternalToken = {uid: 1, host: "example.com", scope: "s"};
const sampleUser: User = {UID: 1, Login: "u", Betas: [], BetaRegistered: false} as any;
const sampleEmails: EmailAddr[] = [{Email: "a@a.com"}] as any;

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
					IncludedEmails: sampleEmails,
				})});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantAuthInfo("t"));
			})).to.eql([
				new UserActions.FetchedUser(sampleUser.UID, sampleUser),
				new UserActions.FetchedAuthInfo("t", sampleAuthInfo),
				new UserActions.FetchedEmails(sampleUser.UID, sampleEmails),
				new UserActions.FetchedGitHubToken(sampleUser.UID, sampleToken),
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
	describe("should handle WantUser", () => {
		it("with user available", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/users/1$");
				return immediateSyncPromise({status: 200, json: () => sampleUser});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantUser(1));
			})).to.eql([new UserActions.FetchedUser(1, sampleUser)]);
		});
		it("with user not available", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/users/1$");
				return immediateSyncPromise({status: 404, text: () => immediateSyncPromise("error", true)});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantUser(1));
			})).to.eql([]);
		});
	});
	describe("should handle WantEmails", () => {
		it("with emails available", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/users/1$/emails");
				return immediateSyncPromise({status: 200, json: () => ({EmailAddrs: sampleEmails})});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantEmails(1));
			})).to.eql([new UserActions.FetchedEmails(1, sampleEmails)]);
		});
		it("with emails not available", () => {
			UserBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/users/1$/emails");
				return immediateSyncPromise({status: 404, text: () => immediateSyncPromise("error", true)});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				UserBackend.__onDispatch(new UserActions.WantEmails(1));
			})).to.eql([]);
		});
	});
});
