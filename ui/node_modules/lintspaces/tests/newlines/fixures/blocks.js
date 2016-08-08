(function() {
	var foo = 'bar';
	var index = 0;
	while (index < foo.length) {
		console.log(foo.charAt(index));
		index++;
	}



	index = foo.length - 1;
	while (index > -1) {
		console.log(foo.charAt(index));
		index--;
	}




	index = 0;
	while (index < foo.length) {
		console.log(foo.charAt(index));
		index++;
	}
});
