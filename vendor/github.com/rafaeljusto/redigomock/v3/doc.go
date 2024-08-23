// Copyright 2014 Rafael Dantas Justo. All rights reserved.

// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// Package redigomock is a mock for redigo library (redis client)
//
// Redigomock basically register the commands with the expected results in a internal global
// variable. When the command is executed via Conn interface, the mock will look to this global
// variable to retrieve the corresponding result.
//
// To start a mocked connection just do the following:
//
//  c := redigomock.NewConn()
//
// Now you can inject it whenever your system needs a redigo.Conn because it satisfies all interface
// requirements. Before running your tests you need beyond of mocking the connection, registering
// the expected results. For that you can generate commands with the expected results.
//
//  c.Command("HGETALL", "person:1").Expect("Person!")
//  c.Command(
//    "HMSET", []string{"person:1", "name", "John"},
//  ).Expect("ok")
//
// As the Expect method from Command receives anything (interface{}), another method was created to
// easy map the result to your structure. For that use ExpectMap:
//
//  c.Command("HGETALL", "person:1").ExpectMap(map[string]string{
//    "name": "John",
//    "age": 42,
//  })
//
// You should also test the error cases, and you can do it in the same way of a normal result.
//
//  c.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Low level error!"))
//
// Sometimes you will want to register a command regardless the arguments, and you can do it with
// the method GenericCommand (mainly with the HMSET).
//
//  c.GenericCommand("HMSET").Expect("ok")
//
// All commands are registered in a global variable, so they will be there until all your test cases
// ends. So for good practice in test writing you should in the beginning of each test case clear
// the mock states.
//
//  c.Clear()
//
// Let's see a full test example. Imagine a Person structure and a function that pick up this
// person in Redis using redigo library (file person.go):
//
//  package person
//
//  import (
//    "fmt"
//    "github.com/gomodule/redigo/redis"
//  )
//
//  type Person struct {
//    Name string `redis:"name"`
//    Age  int    `redis:"age"`
//  }
//
//  func RetrievePerson(conn redis.Conn, id string) (Person, error) {
//    var person Person
//
//    values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
//    if err != nil {
//      return person, err
//    }
//
//    err = redis.ScanStruct(values, &person)
//    return person, err
//  }
//
// Now we need to test it, so let's create the corresponding test with redigomock
// (fileperson_test.go):
//
//  package person
//
//  import (
//    "github.com/rafaeljusto/redigomock/v3"
//    "testing"
//  )
//
//  func TestRetrievePerson(t *testing.T) {
//    conn := redigomock.NewConn()
//    cmd := conn.Command("HGETALL", "person:1").ExpectMap(map[string]string{
//      "name": "Mr. Johson",
//      "age":  "42",
//    })
//
//    person, err := RetrievePerson(conn, "1")
//    if err != nil {
//      t.Fatal(err)
//    }
//
//    if conn.Stats(cmd) != 1 {
//      t.Fatal("Command was not called!")
//    }
//
//    if person.Name != "Mr. Johson" {
//      t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
//    }
//
//    if person.Age != 42 {
//      t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
//    }
//  }
//
//  func TestRetrievePersonError(t *testing.T) {
//    conn := redigomock.NewConn()
//    conn.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Simulate error!"))
//
//    person, err = RetrievePerson(conn, "1")
//    if err == nil {
//      t.Error("Should return an error!")
//    }
//  }
//
// When you use redis as a persistent list, then you might want to call the
// same redis command multiple times. For example:
//
//  func PollForData(conn redis.Conn) error {
//    var url string
//    var err error
//
//    for {
//      if url, err = conn.Do("LPOP", "URLS"); err != nil {
//        return err
//      }
//
//      go func(input string) {
//        // do something with the input
//      }(url)
//    }
//
//    panic("Shouldn't be here")
//  }
//
// To test it, you can chain redis responses. Let's write a test case:
//
//  func TestPollForData(t *testing.T) {
//    conn := redigomock.NewConn()
//    conn.Command("LPOP", "URLS").
//      Expect("www.some.url.com").
//      Expect("www.another.url.com").
//      ExpectError(redis.ErrNil)
//
//    if err := PollForData(conn); err != redis.ErrNil {
//      t.Error("This should return redis nil Error")
//    }
//  }
//
// In the first iteration of the loop redigomock would return
// "www.some.url.com", then "www.another.url.com" and finally redis.ErrNil.
//
// Sometimes providing expected arguments to redigomock at compile time could
// be too constraining. Let's imagine you use redis hash sets to store some
// data, along with the timestamp of the last data update. Let's expand our
// Person struct:
//
//  type Person struct {
//    Name      string `redis:"name"`
//    Age       int    `redis:"age"`
//    UpdatedAt uint64 `redis:updatedat`
//    Phone     string `redis:phone`
//  }
//
// And add a function updating personal data (phone number for example).
// Please notice that the update timestamp can't be determined at compile time:
//
//  func UpdatePersonalData(conn redis.Conn, id string, person Person) error {
//    _, err := conn.Do("HMSET", fmt.Sprint("person:", id), "name", person.Name, "age", person.Age, "updatedat" , time.Now.Unix(), "phone" , person.Phone)
//    return err
//  }
//
// Unit test:
//
//  func TestUpdatePersonalData(t *testing.T){
//    redigomock.Clear()
//
//    person := Person{
//      Name  : "A name",
//      Age   : 18
//      Phone : "123456"
//    }
//
//    conn := redigomock.NewConn()
//    conn.Commmand("HMSET", "person:1", "name", person.Name, "age", person.Age, "updatedat", redigomock.NewAnyInt(), "phone", person.Phone).Expect("OK!")
//
//    err := UpdatePersonalData(conn, "1", person)
//    if err != nil {
//      t.Error("This shouldn't return any errors")
//    }
//  }
//
// As you can see at the position of current timestamp redigomock is told to
// match AnyInt struct created by NewAnyInt() method. AnyInt struct will match
// any integer passed to redigomock from the tested method. Please see
// fuzzyMatch.go file for more details.
//
// The interface of Conn which matches redigo.Conn is safe for concurrent use,
// but the mock-only methods and fields, like Command and Errors, should not be
// accessed concurrently with such calls.
package redigomock
