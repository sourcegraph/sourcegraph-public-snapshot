#include <string>
#include <iostream>
#include <assert.h>
#include "../subset_map.hpp"

Subset_Map<string, string> ssm;

string toString(vector<string> v);
string toString(vector<pair<string, vector<string>>> v);
void assertEqual(string sExpected, string sResult);

void setup() {
  ssm.clear();
  
  //@ssm[Set[1, 2]] = "Foo"
  vector<string> s1;
  s1.push_back("1");
  s1.push_back("2");
  ssm.put(s1, "Foo");
  
  //@ssm[Set["fizz", "fazz"]] = "Bar"
  vector<string> s2;
  s2.push_back("fizz");
  s2.push_back("fazz");
  ssm.put(s2, "Bar");
  
  //@ssm[Set[:foo, :bar]] = "Baz"
  vector<string> s3;
  s3.push_back(":foo");
  s3.push_back(":bar");
  ssm.put(s3, "Baz");
  
  //@ssm[Set[:foo, :bar, :baz]] = "Bang"
  vector<string> s4;
  s4.push_back(":foo");
  s4.push_back(":bar");
  s4.push_back(":baz");
  ssm.put(s4, "Bang");
  
  //@ssm[Set[:bip, :bop, :blip]] = "Qux"
  vector<string> s5;
  s5.push_back(":bip");
  s5.push_back(":bop");
  s5.push_back(":blip");
  ssm.put(s5, "Qux");
  
  //@ssm[Set[:bip, :bop]] = "Thram"
  vector<string> s6;
  s6.push_back(":bip");
  s6.push_back(":bop");
  ssm.put(s6, "Thram");
}

void testEqualKeys() {
  cout << "testEqualKeys" << endl;
  
  //assert_equal [["Foo", Set[1, 2]]], @ssm.get(Set[1, 2])
  vector<string> k1;
  k1.push_back("1");
  k1.push_back("2");
  assertEqual("[[Foo, Set[1, 2]]]", toString(ssm.get_kv(k1)));
  
  //assert_equal [["Bar", Set["fizz", "fazz"]]], @ssm.get(Set["fizz", "fazz"])
  vector<string> k2;
  k2.push_back("fizz");
  k2.push_back("fazz");
  assertEqual("[[Bar, Set[fizz, fazz]]]", toString(ssm.get_kv(k2)));
  
  cout << endl;
}

void testSubsetKeys() {
  cout << "testSubsetKeys" << endl;
  
  //assert_equal [["Foo", Set[1, 2]]], @ssm.get(Set[1, 2, "fuzz"])
  vector<string> k1;
  k1.push_back("1");
  k1.push_back("2");
  k1.push_back("fuzz");
  assertEqual("[[Foo, Set[1, 2]]]", toString(ssm.get_kv(k1)));
  
  //assert_equal [["Bar", Set["fizz", "fazz"]]], @ssm.get(Set["fizz", "fazz", 3])
  vector<string> k2;
  k2.push_back("fizz");
  k2.push_back("fazz");
  k2.push_back("3");
  assertEqual("[[Bar, Set[fizz, fazz]]]", toString(ssm.get_kv(k2)));
  
  cout << endl;
}

void testSupersetKeys() {
  cout << "testSupersetKeys" << endl;
  
  //assert_equal [], @ssm.get(Set[1])
  vector<string> k1;
  k1.push_back("1");
  assertEqual("[]", toString(ssm.get_kv(k1)));
  
  //assert_equal [], @ssm.get(Set[2])
  vector<string> k2;
  k2.push_back("2");
  assertEqual("[]", toString(ssm.get_kv(k2)));
  
  //assert_equal [], @ssm.get(Set["fizz"])
  vector<string> k3;
  k3.push_back("fizz");
  assertEqual("[]", toString(ssm.get_kv(k3)));
  
  //assert_equal [], @ssm.get(Set["fazz"])
  vector<string> k4;
  k4.push_back("fazz");
  assertEqual("[]", toString(ssm.get_kv(k4)));
  
  cout << endl;
}

void testDisjointKeys() {
  cout << "testDisjointKeys" << endl;
  
  //assert_equal [], @ssm.get(Set[3, 4])
  vector<string> k1;
  k1.push_back("3");
  k1.push_back("4");
  assertEqual("[]", toString(ssm.get_kv(k1)));
  
  //assert_equal [], @ssm.get(Set["fuzz", "frizz"])
  vector<string> k2;
  k2.push_back("fuzz");
  k2.push_back("frizz");
  assertEqual("[]", toString(ssm.get_kv(k2)));
  
  //assert_equal [], @ssm.get(Set["gran", 15])
  vector<string> k3;
  k3.push_back("gran");
  k3.push_back("15");
  assertEqual("[]", toString(ssm.get_kv(k3)));
  
  cout << endl;
}

void testSemiDisjointKeys() {
  cout << "testSemiDisjointKeys" << endl;
  
  //assert_equal [], @ssm.get(Set[2, 3])
  vector<string> k1;
  k1.push_back("2");
  k1.push_back("3");
  assertEqual("[]", toString(ssm.get_kv(k1)));
  
  //assert_equal [], @ssm.get(Set["fizz", "fuzz"])
  vector<string> k2;
  k2.push_back("fizz");
  k2.push_back("fuzz");
  assertEqual("[]", toString(ssm.get_kv(k2)));
  
  //assert_equal [], @ssm.get(Set[1, "fazz"])
  vector<string> k3;
  k3.push_back("1");
  k3.push_back("fazz");
  assertEqual("[]", toString(ssm.get_kv(k3)));
  
  cout << endl;
}

void testEmptyKeySet() {
  cout << "testEmptyKeySet" << endl;
  
  //assert_raises(ArgumentError) {@ssm[Set[]] = "Fail"}
  vector<string> s1;
  try {
    ssm.put(s1, "Fail");
  }
  catch (const char* &e) {
    assertEqual("internal error: subset map keys may not be empty", e);
  }
}

void testEmptyKeyGet() {
  cout << "testEmptyKeyGet" << endl;
  
  //assert_equal [], @ssm.get(Set[])
  vector<string> k1;
  assertEqual("[]", toString(ssm.get_kv(k1)));
  
  cout << endl;
}
void testMultipleSubsets() {
  cout << "testMultipleSubsets" << endl;
  
  //assert_equal [["Foo", Set[1, 2]], ["Bar", Set["fizz", "fazz"]]], @ssm.get(Set[1, 2, "fizz", "fazz"])
  vector<string> k1;
  k1.push_back("1");
  k1.push_back("2");
  k1.push_back("fizz");
  k1.push_back("fazz");
  assertEqual("[[Foo, Set[1, 2]], [Bar, Set[fizz, fazz]]]", toString(ssm.get_kv(k1)));
  
  //assert_equal [["Foo", Set[1, 2]], ["Bar", Set["fizz", "fazz"]]], @ssm.get(Set[1, 2, 3, "fizz", "fazz", "fuzz"])
  vector<string> k2;
  k2.push_back("1");
  k2.push_back("2");
  k2.push_back("3");
  k2.push_back("fizz");
  k2.push_back("fazz");
  k2.push_back("fuzz");
  assertEqual("[[Foo, Set[1, 2]], [Bar, Set[fizz, fazz]]]", toString(ssm.get_kv(k2)));
  
  //assert_equal [["Baz", Set[:foo, :bar]]], @ssm.get(Set[:foo, :bar])
  vector<string> k3;
  k3.push_back(":foo");
  k3.push_back(":bar");
  assertEqual("[[Baz, Set[:foo, :bar]]]", toString(ssm.get_kv(k3)));
  
  //assert_equal [["Baz", Set[:foo, :bar]], ["Bang", Set[:foo, :bar, :baz]]], @ssm.get(Set[:foo, :bar, :baz])
  vector<string> k4;
  k4.push_back(":foo");
  k4.push_back(":bar");
  k4.push_back(":baz");
  assertEqual("[[Baz, Set[:foo, :bar]], [Bang, Set[:foo, :bar, :baz]]]", toString(ssm.get_kv(k4)));
  
  cout << endl;
}
void testBracketBracket() {
  cout << "testBracketBracket" << endl;
  
  //assert_equal ["Foo"], @ssm[Set[1, 2, "fuzz"]]
  vector<string> k1;
  k1.push_back("1");
  k1.push_back("2");
  k1.push_back("fuzz");
  assertEqual("[Foo]", toString(ssm.get_v(k1)));
  
  //assert_equal ["Baz", "Bang"], @ssm[Set[:foo, :bar, :baz]]
  vector<string> k2;
  k2.push_back(":foo");
  k2.push_back(":bar");
  k2.push_back(":baz");
  assertEqual("[Baz, Bang]", toString(ssm.get_v(k2)));
  
  cout << endl;
}

void testKeyOrder() {
  cout << "testEqualKeys" << endl;
  
  //assert_equal [["Foo", Set[1, 2]]], @ssm.get(Set[2, 1])
  vector<string> k1;
  k1.push_back("2");
  k1.push_back("1");
  assertEqual("[[Foo, Set[1, 2]]]", toString(ssm.get_kv(k1)));
  
  cout << endl;
}

void testOrderPreserved() {
  cout << "testOrderPreserved" << endl;
  //@ssm[Set[10, 11, 12]] = 1
  vector<string> s1;
  s1.push_back("10");
  s1.push_back("11");
  s1.push_back("12");
  ssm.put(s1, "1");
  
  //@ssm[Set[10, 11]] = 2
  vector<string> s2;
  s2.push_back("10");
  s2.push_back("11");
  ssm.put(s2, "2");
  
  //@ssm[Set[11]] = 3
  vector<string> s3;
  s3.push_back("11");
  ssm.put(s3, "3");
  
  //@ssm[Set[11, 12]] = 4
  vector<string> s4;
  s4.push_back("11");
  s4.push_back("12");
  ssm.put(s4, "4");
  
  //@ssm[Set[9, 10, 11, 12, 13]] = 5
  vector<string> s5;
  s5.push_back("9");
  s5.push_back("10");
  s5.push_back("11");
  s5.push_back("12");
  s5.push_back("13");
  ssm.put(s5, "5");
  
  //@ssm[Set[10, 13]] = 6
  vector<string> s6;
  s6.push_back("10");
  s6.push_back("13");
  ssm.put(s6, "6");
  
  //assert_equal([[1, Set[10, 11, 12]], [2, Set[10, 11]], [3, Set[11]], [4, Set[11, 12]], [5, Set[9, 10, 11, 12, 13]], [6, Set[10, 13]]], @ssm.get(Set[9, 10, 11, 12, 13]))
  vector<string> k1;
  k1.push_back("9");
  k1.push_back("10");
  k1.push_back("11");
  k1.push_back("12");
  k1.push_back("13");
  assertEqual("[[1, Set[10, 11, 12]], [2, Set[10, 11]], [3, Set[11]], [4, Set[11, 12]], [5, Set[9, 10, 11, 12, 13]], [6, Set[10, 13]]]", toString(ssm.get_kv(k1)));
  
  cout << endl;
}
void testMultipleEqualValues() {
  cout << "testMultipleEqualValues" << endl;
  //@ssm[Set[11, 12]] = 1
  vector<string> s1;
  s1.push_back("11");
  s1.push_back("12");
  ssm.put(s1, "1");
  
  //@ssm[Set[12, 13]] = 2
  vector<string> s2;
  s2.push_back("12");
  s2.push_back("13");
  ssm.put(s2, "2");
  
  //@ssm[Set[13, 14]] = 1
  vector<string> s3;
  s3.push_back("13");
  s3.push_back("14");
  ssm.put(s3, "1");
  
  //@ssm[Set[14, 15]] = 1
  vector<string> s4;
  s4.push_back("14");
  s4.push_back("15");
  ssm.put(s4, "1");
  
  //assert_equal([[1, Set[11, 12]], [2, Set[12, 13]], [1, Set[13, 14]], [1, Set[14, 15]]], @ssm.get(Set[11, 12, 13, 14, 15]))
  vector<string> k1;
  k1.push_back("11");
  k1.push_back("12");
  k1.push_back("13");
  k1.push_back("14");
  k1.push_back("15");
  assertEqual("[[1, Set[11, 12]], [2, Set[12, 13]], [1, Set[13, 14]], [1, Set[14, 15]]]", toString(ssm.get_kv(k1)));
  
  cout << endl;
}

int main()
{
  vector<string> s1;
  s1.push_back("1");
  s1.push_back("2");
  
  vector<string> s2;
  s2.push_back("2");
  s2.push_back("3");
  
  vector<string> s3;
  s3.push_back("3");
  s3.push_back("4");
  
  ssm.put(s1, "value1");
  ssm.put(s2, "value2");
  ssm.put(s3, "value3");
  
  vector<string> s4;
  s4.push_back("1");
  s4.push_back("2");
  s4.push_back("3");
  
  vector<pair<string, vector<string> > > fetched(ssm.get_kv(s4));
  
  cout << "PRINTING RESULTS:" << endl;
  for (size_t i = 0, S = fetched.size(); i < S; ++i) {
    cout << fetched[i].first << endl;
  }
  
  Subset_Map<string, string> ssm2;
  ssm2.put(s1, "foo");
  ssm2.put(s2, "bar");
  ssm2.put(s4, "hux");
  
  vector<pair<string, vector<string> > > fetched2(ssm2.get_kv(s4));
  
  cout << endl << "PRINTING RESULTS:" << endl;
  for (size_t i = 0, S = fetched2.size(); i < S; ++i) {
    cout << fetched2[i].first << endl;
  }
  
  cout << "TRYING ON A SELECTOR-LIKE OBJECT" << endl;
  
  Subset_Map<string, string> sel_ssm;
  vector<string> target;
  target.push_back("desk");
  target.push_back(".wood");
  
  vector<string> actual;
  actual.push_back("desk");
  actual.push_back(".wood");
  actual.push_back(".mine");
  
  sel_ssm.put(target, "has-aquarium");
  vector<pair<string, vector<string> > > fetched3(sel_ssm.get_kv(actual));
  cout << "RESULTS:" << endl;
  for (size_t i = 0, S = fetched3.size(); i < S; ++i) {
    cout << fetched3[i].first << endl;
  }
  
  cout << endl;
  
  // BEGIN PORTED RUBY TESTS FROM /test/sass/util/subset_map_test.rb
  
  setup();
  testEqualKeys();
  testSubsetKeys();
  testSupersetKeys();
  testDisjointKeys();
  testSemiDisjointKeys();
  testEmptyKeySet();
  testEmptyKeyGet();
  testMultipleSubsets();
  testBracketBracket();
  testKeyOrder();
  
  setup();
  testOrderPreserved();
  
  setup();
  testMultipleEqualValues();
  
  return 0;
}

string toString(vector<pair<string, vector<string>>> v)
{
  stringstream buffer;
  buffer << "[";
  for (size_t i = 0, S = v.size(); i < S; ++i) {
    buffer << "[" << v[i].first;
    buffer << ", Set[";
    for (size_t j = 0, S = v[i].second.size(); j < S; ++j) {
      buffer << v[i].second[j];
      if (j < S-1) {
        buffer << ", ";
      }
    }
    buffer << "]]";
    if (i < S-1) {
      buffer << ", ";
    }
  }
  buffer << "]";
  return buffer.str();
}

string toString(vector<string> v)
{
  stringstream buffer;
  buffer << "[";
  for (size_t i = 0, S = v.size(); i < S; ++i) {
    buffer << v[i];
    if (i < S-1) {
      buffer << ", ";
    }
  }
  buffer << "]";
  return buffer.str();
}

void assertEqual(string sExpected, string sResult) {
  cout << "Expected: " << sExpected << endl;
  cout << "Result:   " << sResult << endl;
  assert(sExpected == sResult);
}
