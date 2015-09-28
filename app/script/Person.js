exports.label = function(person) {
	if (person.Login) return person.Login;
	return person.Email.replace(/@.+$/, "@…");
};
