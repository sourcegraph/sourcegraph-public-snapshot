$(document).ready(function(){
  module('String.score');
  
    test('Exact match', function(){
      expect(1);
      equals('Hello World'.score('Hello World'), 1.0, 'Exact matches should score 1.0');
    });
  
    test('Not matchhing', function(){
      expect(2);
      equals("hello world".score("hellx"), 0, 'non-existint charactor in match should return 0');
      equals("hello world".score("hello_world"),0, 'non-existint charactor in match should return 0');
    });
  
    test('Match must be sequential', function(){
      equals('Hello World'.score('WH'), 0, 'Matches out of order should return 0');
    });
  
    test('Same case should match better then wrong case', function(){
      ok('Hello World'.score('hello')<'Hello World'.score('Hello'));
    });
  
    test('Higher score for closer matches', function(){
      ok('Hello World'.score('H')<'Hello World'.score('He'), '"He" should match "Hello World" better then "H" does');
    });
  
    test('Matching with wrong case', function(){
      ok("Hillsdale Michigan".score("himi")>0, 'should match first matchable letter regardless of case');
    });
  
    test('should have proper relative weighting', function(){
      ok( "hello world".score("e")          < "hello world".score("h") );
      ok( "hello world".score("h")          < "hello world".score("he") );
      ok( "hello world".score("hel")        < "hello world".score("hell") );
      ok( "hello world".score("hell")       < "hello world".score("hello") );
      ok( "hello world".score("hello")      < "hello world".score("helloworld") );
      ok( "hello world".score("helloworl")  < "hello world".score("hello worl") );
      ok( "hello world".score("hello worl") < "hello world".score("hello world") );
    });
  
  module('Advanced Scoreing Methods');
    test('consecutive letter bonus', function(){
      expect(1);
      ok('Hello World'.score('Hel') > 'Hello World'.score('Hld'), '"Hel" should match "Hello World" better then "Hld"');
    });
  
    test('Acronym bonus', function(){
      expect(6);
      ok('Hello World'.score('HW') > 'Hello World'.score('Ho'), '"HW" should score higher with "Hello World" then Ho');
      ok('yet another Hello World'.score('yaHW') > 'Hello World'.score('yet another'));
      ok("Hillsdale Michigan".score("HiMi") > "Hillsdale Michigan".score("Hil"), '"HiMi" should match "Hillsdale Michigan" higher then "Hil"');
      ok("Hillsdale Michigan".score("HiMi") > "Hillsdale Michigan".score("illsda"));
      ok("Hillsdale Michigan".score("HiMi") > "Hillsdale Michigan".score("hills")); 
      ok("Hillsdale Michigan".score("HiMi") < "Hillsdale Michigan".score("hillsd")); 
    });
  
    test('Beginning of string bonus', function(){
      expect(2);
      ok("Hillsdale".score("hi") > "Chippewa".score("hi"));
      ok("hello world".score("h") > "hello world".score("w"), 'should have a bonus for matching first letter');
    });
  
    test('proper string weights', function(){
      ok("Research Resources North".score('res') > "Mary Conces".score('res'), 'res matches "Mary Conces" better then "Research Resources North"');
    
      ok("Research Resources North".score('res') > "Bonnie Strathern - Southwest Michigan Title Search".score('res'));
    });
  
    test('Start of String bonus', function(){
      ok("Mary Large".score('mar') > "Large Mary".score('mar'));
      ok("Silly Mary Large".score('mar') === "Silly Large Mary".score('mar')); // ensure start of string bonus only on start of string
    });
  
  module('Fuzzy String');
    test('should score mismatched strings', function(){
      expect(2);
      equal('Hello World'.score('Hz'), 0, 'should score 0 without a specified fuzziness.');
      ok('Hello World'.score('Hz', 0.5) < 'Hello World'.score('H', 0.5), 'fuzzy matches should be worse then good ones');
    });
    
    test('should be tuned well', function(){
      expect(2);
      ok("hello world".score("hello worl", 0.5) > "hello world".score("hello wor1", 0.5), 'mismatch should always be worse');
      ok('Hello World'.score('jello',0.5) > 0, '"Hello World" should match "jello" more then 0 with a fuzziness of 0.5');
    });
    
    test('should have varying degrees of fuzziness', function(){
      expect(1);
      ok('Hello World'.score('Hz', 0.9) > 'Hello World'.score('Hz', 0.5), 'higher fuzziness should yield higher scores');
    });
  module('Benchmark');
    test('Expand to see time to score', function(){
      var iterations = 4000;
      
      var start1 = new Date().valueOf();
      for(i=iterations;i>0;i--){ "hello world".score("h"); }
      var end1 = new Date().valueOf();
      var t1=end1-start1;
      ok(true, t1 + ' miliseconds to do '+iterations+' iterations of "hello world".score("h")');
      
      var start2 = new Date().valueOf();
      for(i=iterations;i>0;i--){ "hello world".score("hw"); }
      var end2 = new Date().valueOf();
      var t2=end2-start2;
      ok(true, t2 + ' miliseconds to do '+iterations+' iterations of "hello world".score("hw")');
      
      var start3 = new Date().valueOf();
      for(i=iterations;i>0;i--){ "hello world".score("hello world"); }
      var end3 = new Date().valueOf();
      var t3=end3-start3;
      ok(true, t3 + ' miliseconds to do '+iterations+' iterations of "hello world".score("hello world")');
      
      var start4 = new Date().valueOf();
      for(i=iterations;i>0;i--){ "hello any world that will listen".score("hlo wrdthlstn"); }
      var end4 = new Date().valueOf();
      var t4=end4-start4;
      ok(true, t4 + ' miliseconds to do '+iterations+' iterations of "hello any world that will listen".score("hlo wrdthlstn")');
      
      var start5 = new Date().valueOf();
      for(i=iterations;i>0;i--){ "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.".score("Lorem i dor coecadipg et, Duis aute irure dole nulla. qui ofa mot am l"); }
      var end5 = new Date().valueOf();
      var t5=end5-start5;
      ok(true, t5 + ' miliseconds to do '+iterations+' iterations of 446 character string scoring a 70 character match');
      
      ok(true, 'Average Score... all benchmarks devided by the number of benchmarks (smaller is faster): '+ (t1+t2+t3+t4+t5)/5) + 'ms';
    });
});