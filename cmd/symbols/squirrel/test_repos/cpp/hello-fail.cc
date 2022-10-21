#include "examples/cpp/hello-lib.h"

#include <string>

using hello::HelloLib;
using std::string;

/**
 * This is a fake test that prints "Hello barf" and then exits with exit code 1.
 * If run as a test (cc_test), the non-0 exit code indicates to Bazel that the
 * "test" has failed.
 */
//  vvvv fail-main def
//           vvvv fail-main.argc def
//                        vvvv fail-main.argv def
int main(int argc, char** argv) {
//         vvv fail-main.thing ref
  HelloLib lib("Hello");
//       vvvvv fail-main.thing def
  string thing = "barf";
//    vvvv fail-main.argc ref
  if (argc > 1) {
//  vvvvv fail-main.thing ref
//          vvvv fail-main.argv ref
    thing = argv[1];
  }

  lib.greet(thing);
  return 1;
}
