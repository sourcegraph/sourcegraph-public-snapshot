#ifndef EXAMPLES_CPP_HELLO_LIB_H_ // < "#" hellolib.h def
#define EXAMPLES_CPP_HELLO_LIB_H_

#include <string>
#include <memory>

namespace hello {

class HelloLib {
 public:
  explicit HelloLib(const std::string &greeting);

  void greet(const std::string &thing);

 private:
  std::unique_ptr<const std::string> greeting_;
};

}  // namespace hello

#endif  // EXAMPLES_CPP_HELLO_LIB_H_
