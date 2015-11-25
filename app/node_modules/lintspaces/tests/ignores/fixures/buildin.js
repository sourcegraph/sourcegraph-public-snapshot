/**
 * This is a multiline comment which should be INGORED.
 * It contains extra trailingspaces.   
 */
(function() {
	var foo = 'bar';
	var index = 0;

	/**
	 * This is a multiline comment which should be INGORED.
	 * It contains extra trailingspaces. 
	 */
	while (index < foo.length) {
		console.log(foo.charAt(index));

		 /* This is a multiline comment which SHOULD FAIL BECAUSE OF A WRONG
		 * INDENTION AT THE BEGINNING OF THE FIRST LINE OF THIS COMMENT.
		 *
		 * It contains extra trailingspaces.        
		 */
		index++;
	}
});


/**
 * This is a multiline comment which should be INGORED.
 * It contains extra trailingspaces.   
 */
while (index < foo.length) {
	console.log(foo.charAt(index));
	index++;
}
