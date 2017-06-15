var test = {
	name: "Login",
	timeout: 20,
	steps: [
		{action: "navigate", url: "http://localhost:3080/"},
		{action: "click", selector: {text: "Sign in"}},
		{action: "click", selector: {text: "Sign in", background: "light"}},
		{action: "click", selector: {text: "Username"}},
		{action: "type", text: "sg-e2e-chrome-test"},
		{action: "click", selector: {text: "Password"}},
		{action: "type", text: "PJLRQK8EUhvTozRit7RNDZjO"},
		{action: "click", selector: {text: "Sign in", background: "dark"}},
		{action: "click", selector: {text: "Authorize sourcegraph", attributes: {disabled: null}}},
		{action: "find", selector: {text: "Welcome!"}},
	],
}

console.log(JSON.stringify(test));
