// Syntax-highlight all code blocks in the readme on the client.
document.addEventListener("DOMContentLoaded", function() {
	var fileBody = document.querySelector(".plain-file .panel-body");
	if (!fileBody) return;

	// If the entire readme is a plain text file that we've rendered with a single <pre>, treat it as prose, not code (and don't syntax-highlight it.
	if (fileBody.children.length === 1) return;

	var codeElems = document.querySelectorAll(".plain-file .panel-body pre");
	for (var i = 0; i < codeElems.length; i++) {
		codeElems[i].classList.add("prettyprint");
	}
	window.prettyPrint(function() {}, fileBody);
});
