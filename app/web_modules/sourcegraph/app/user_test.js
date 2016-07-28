// @flow weak

import * as React from "react";
import expect from "expect.js";
import {withUserContext, getChildContext} from "sourcegraph/app/user";
import {render} from "sourcegraph/util/renderTestUtils";
import UserStore from "sourcegraph/user/UserStore";
import * as UserActions from "sourcegraph/user/UserActions";
import type {User, AuthInfo, ExternalToken} from "sourcegraph/user";

const sampleAuthInfo: AuthInfo = {UID: 1, Login: "u"};
const sampleGitHubToken: ExternalToken = {uid: 1, host: "example.com", scope: "s"};
const sampleUser: User = {UID: 1, Login: "u", Betas: [], BetaRegistered: false};

const C = withUserContext((props) => null);
const renderAndGetContext = (c) => {
	const res = render(c, {});

	// Hack to get state, so we can pass it to getChildContext.
	const state = {};
	const e = new C({});
	e.reconcileState(state, {});
	e.onStateTransition(state, state);
	return {...res, context: getChildContext(state)};
};

describe("withUserContext", () => {
	it("no accessToken", () => {
		UserStore.activeAccessToken = null;
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([]);
		expect(res.context).to.eql({authInfo: null, user: null, signedIn: false, githubToken: null});
	});
	it("with accessToken, no authInfo yet", () => {
		UserStore.activeAccessToken = "t";
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([new UserActions.WantAuthInfo("t")]);
		expect(res.context).to.eql({authInfo: null, user: null, signedIn: true, githubToken: null});
	});
	it("with accessToken and authInfo fetched, no user yet", () => {
		UserStore.activeAccessToken = "t";
		UserStore.directDispatch(new UserActions.FetchedAuthInfo("t", sampleAuthInfo));
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([new UserActions.WantUser(1)]);
		expect(res.context).to.eql({authInfo: {Login: "u", UID: 1}, user: null, signedIn: true, githubToken: null});
	});
	it("with accessToken, authInfo, and user", () => {
		UserStore.activeAccessToken = "t";
		UserStore.directDispatch(new UserActions.FetchedAuthInfo("t", sampleAuthInfo));
		UserStore.directDispatch(new UserActions.FetchedUser(1, sampleUser));
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([]);
		expect(res.context).to.eql({authInfo: {Login: "u", UID: 1}, user: sampleUser, signedIn: true, githubToken: null});
	});
	it("with accessToken but empty authInfo object (indicating no user, expired accessToken, etc.)", () => {
		UserStore.activeAccessToken = "t";
		UserStore.directDispatch(new UserActions.FetchedAuthInfo("t", {}));
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([]);
		expect(res.context.signedIn).to.be(false);
	});
	it("with GitHub token", () => {
		UserStore.activeGitHubToken = sampleGitHubToken;
		const res = renderAndGetContext(<C />);
		expect(res.actions).to.eql([]);
		expect(res.context).to.eql({authInfo: null, user: null, signedIn: false, githubToken: sampleGitHubToken});
	});
});
