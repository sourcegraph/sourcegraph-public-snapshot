/**
 * This is a multiline comment which should be INGORED.
 * It contains extra trailingspaces.   
 */
(function() {
	var foo = 'bar';
	var index = 0;

	/** This is a multline comment in a single line */      

	/** This is a multline comment in a single line */      
	/**   
	 * This is a multiline comment   
	 * It contains extra trailingspaces. 
	 */       

	/** This is a multline comment in a single line */      
	/**
	 * This is a multiline comment   
	 * It contains extra trailingspaces. 
	 */       

	/**   
	 * This is a multiline comment   
	 * It contains extra trailingspaces. 
	 */       
	while (index < foo.length) {
		console.log(foo.charAt(index));

		/* This is a multiline comment which including trailingspaces    
		 * It contains extra trailingspaces.        
		 */
		index++;
	}
});
