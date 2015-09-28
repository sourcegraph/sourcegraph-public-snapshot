#include <assert.h>
#include <sstream>

#include "node.hpp"
#include "to_string.hpp"
#include "parser.hpp"


#define STATIC_ARRAY_SIZE(array) (sizeof((array))/sizeof((array[0])))


namespace Sass {
  
  Context ctx = Context::Data();
  
  To_String to_string;
  
  
  const char* const ROUNDTRIP_TESTS[] = {
    NULL,
  	"~",
		"CMPD",
    "~ CMPD",
    "CMPD >",
    "> > CMPD",
    "CMPD ~ ~",
    "> + CMPD1.CMPD2 > ~",
    "> + CMPD1.CMPD2 CMPD3.CMPD4 > ~",
    "+ CMPD1 CMPD2 ~ CMPD3 + CMPD4 > CMPD5 > ~"
  };

  
  
  static Complex_Selector* createComplexSelector(string src) {
    string temp(src);
    temp += ";";
    return (*Parser::from_c_str(temp.c_str(), ctx, "", Position()).parse_selector_group())[0];
  }
  
  
  void roundtripTest(const char* toTest) {

    // Create the initial selector

    Complex_Selector* pOrigSelector = NULL;
    if (toTest) {
      pOrigSelector = createComplexSelector(toTest);
    }
    
    string expected(pOrigSelector ? pOrigSelector->perform(&to_string) : "NULL");
  
    
    // Roundtrip the selector into a node and back
    
    Node node = complexSelectorToNode(pOrigSelector, ctx);
    
    stringstream nodeStringStream;
    nodeStringStream << node;
    string nodeString = nodeStringStream.str();
    cout << "ASNODE: " << node << endl;
    
    Complex_Selector* pNewSelector = nodeToComplexSelector(node, ctx);
    
    // Show the result

    string result(pNewSelector ? pNewSelector->perform(&to_string) : "NULL");
    
    cout << "SELECTOR: " << expected << endl;
    cout << "NEW SELECTOR:   " << result << endl;

    
    // Test that they are equal using the equality operator
    
    assert( (!pOrigSelector && !pNewSelector ) || (pOrigSelector && pNewSelector) );
    if (pOrigSelector) {
	    assert( *pOrigSelector == *pNewSelector );
    }

    
    // Test that they are equal by comparing the string versions of the selectors

    assert(expected == result);
    
  }


	int main() {
    for (int index = 0; index < STATIC_ARRAY_SIZE(ROUNDTRIP_TESTS); index++) {
      const char* const toTest = ROUNDTRIP_TESTS[index];
      cout << "\nINPUT STRING: " << (toTest ? toTest : "NULL") << endl;
      roundtripTest(toTest);
    }
    
    cout << "\nTesting Done.\n";
  }


}