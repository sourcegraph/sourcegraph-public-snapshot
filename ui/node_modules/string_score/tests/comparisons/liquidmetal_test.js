$(document).ready(function(){
  module('String.score');
  
  test('Exact match', function(){
    expect(1);
    equals(LiquidMetal.score('Hello World', 'Hello World'), 1.0);
  });
  
  test('Not matchhing', function(){
    expect(2);
    equals(LiquidMetal.score("hello world","hellx"), 0, 'non-existint charactor in match should return 0');
    equals(LiquidMetal.score("hello world","hello_world"),0, 'non-existint charactor in match should return 0');
  });
  
  test('Match must be sequential', function(){
    ok(!LiquidMetal.score('Hello World','WH'));
    ok(LiquidMetal.score('Hello World','HW'));
  });
  
  test('Same case should match better then wrong case', function(){
    ok(LiquidMetal.score('Hello World','hello')<LiquidMetal.score('Hello World','Hello'));
  });
  
  test('Closer matches should have higher scores', function(){
    ok(LiquidMetal.score('Hello World','H')<LiquidMetal.score('Hello World','He'));
    ok(LiquidMetal.score('Hello World','H')<LiquidMetal.score('Hello World','He'));
  });
  
  test('should match first matchable letter regardless of case', function(){
    ok(LiquidMetal.score("Hillsdale Michigan","himi")>0);
  });
  
  module('Advanced Scoreing Methods');
  test('consecutive letter bonus', function(){
    expect(1);
    ok(LiquidMetal.score('Hello World','Hel') > LiquidMetal.score('Hello World','Hld'));
  });
  
  test('Acronym bonus', function(){
    expect(5);
    ok(LiquidMetal.score('Hello World','HW') > LiquidMetal.score('Hello World','Ho'), '"HW" should score higher with "Hello World" then Ho');
    ok(LiquidMetal.score('yet another Hello World','yaHW') > LiquidMetal.score('Hello World','yet another'));
    ok(LiquidMetal.score("Hillsdale Michigan","HiMi") > LiquidMetal.score("Hillsdale Michigan","Hil"), '"HiMi" should match "Hillsdale Michigan" higher then "Hil"');
    ok(LiquidMetal.score("Hillsdale Michigan","HiMi") > LiquidMetal.score("Hillsdale Michigan","illsda"));
    ok(LiquidMetal.score("Hillsdale Michigan","HiMi") < LiquidMetal.score("Hillsdale Michigan","hills")); // but not higher then matching start of word
  });
  
  test('Beginning of string bonus', function(){
    expect(1);
    ok(LiquidMetal.score("Hillsdale","hi") > LiquidMetal.score("Chippewa","hi"));
  });
  
  test('proper string weights', function(){
    ok(LiquidMetal.score("Research Resources North",'res') > LiquidMetal.score("Mary Conces",'res'), 'res matches "Mary Conces" better then "Research Resources North"');
    
    ok(LiquidMetal.score("Research Resources North",'res') > LiquidMetal.score("Bonnie Strathern - Southwest Michigan Title Search",'res'));
  });
  
  test('Start of String bonus', function(){
    ok(LiquidMetal.score("Mary Large",'mar') > LiquidMetal.score("Large Mary",'mar'));
    ok(LiquidMetal.score("Silly Mary Large", 'mar') === LiquidMetal.score("Silly Large Mary",'mar')); // ensure start of string bonus only on start of string
  });
  
  module('Benchmark');
      test('Expand to see time to score', function(){
        var iterations = 4000;
      
        var start1 = new Date().valueOf();
        for(i=iterations;i>0;i--){ LiquidMetal.score("hello world", "h"); }
        var end1 = new Date().valueOf();
        var t1=end1-start1;
        ok(true, t1 + ' miliseconds to do '+iterations+' iterations of LiquidMetal.score("hello world","h")');
        
        var start2 = new Date().valueOf();
        for(i=iterations;i>0;i--){ LiquidMetal.score("hello world","hw"); }
        var end2 = new Date().valueOf();
        var t2=end2-start2;
        ok(true, t2 + ' miliseconds to do '+iterations+' iterations of LiquidMetal.score("hello world","hw")');
      
        var start3 = new Date().valueOf();
        for(i=iterations;i>0;i--){ LiquidMetal.score("hello world","hello world"); }
        var end3 = new Date().valueOf();
        var t3=end3-start3;
        ok(true, t3 + ' miliseconds to do '+iterations+' iterations of LiquidMetal.score("hello world","hello world")');
        
        var start4 = new Date().valueOf();
        for(i=iterations;i>0;i--){ LiquidMetal.score("hello any world that will listen","hlo wrdthlstn"); }
        var end4 = new Date().valueOf();
        var t4=end4-start4;
        ok(true, t4 + ' miliseconds to do '+iterations+' iterations of LiquidMetal.score("hello any world that will listen","hlo wrdthlstn")');
      
        var start5 = new Date().valueOf();
        for(i=iterations;i>0;i--){ LiquidMetal.score("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", "Lorem i dor coecadipg et, Duis aute irure dole nulla. qui ofa mot am l"); }
        var end5 = new Date().valueOf();
        var t5=end5-start5;
        ok(true, t5 + ' miliseconds to do '+iterations+' iterations of 446 character string scoring a 70 character match');
        
        ok(true, 'score (smaller is faster): '+ (t1+t2+t3+t4+t5)/5);
      });
});