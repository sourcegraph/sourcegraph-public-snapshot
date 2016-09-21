jaro = function(str1, str2){
  var s1 = str1.length,
      s2 = str2.length,
      mw = Math.max(s1, s2)/2-1,
      t=0,
      m=0,
      letter='',
      c = 0, /* case bonus (0 to the length of max(str1,str2)) */
      s = 0, /* Start of string bonus (0 or 1) */
      b = 0, /* beginning of word bonus (0 to 1*(# of words)) */
      // 
      dj = 0,                   /* distance Jaro */
      w = 0, /* number of words (not a bonus)*/
      isLetterIn = function(a,b){
        return Math.max(b.indexOf(a.toLowerCase()), b.indexOf(a.toUpperCase())) > -1;
      };

      
      
  // Test if swapping strX & lenStrX if stra is longer then str2 for proformance ??
  // another option is to bail out of the stepping once we are outside of the context of the other string
  // the issue is that with string lengths of 11 & 2 you wouldn't want to go through the loop 11 times
  
  /* find matches & transpositions */
  for (var i in str2) {
  // for(var i=0; i<s2; i++) {
    letter = str2[i];
    // console.log(letter);
    // if(typeof(letter) != 'string') break;
    if(
      // isLetterIn(letter, str1.slice(i,i+mw))      
      str1.slice(i,i+mw).indexOf(letter) 
      ) { /* match */
      m++;
    } else if(str1.slice(i-mw,i).indexOf(letter) > -1) { /* transposition */
      m++; t++;
    }
    
    
    // if(str1.slice(i,i+mw).indexOf(letter) > -1) { /* match */
    //   m++;
    // } else if(str1.slice(i-mw,i).indexOf(letter) > -1) { /* transposition */
    //   m++; t++;
    // }
  };
  return (1/3*(m/s1+m/s2+(m-t)/m));
};


// jpJaro = function(str1, str2){
//   var s1 = str1.length,         /* first string width*/
//       s2 = str2.length,         /* second string width */
//       w = Math.max(s1, s2)/2-1, /* match window */
//       t=0,                      /* transpositions */
//       m=0,                      /* matches */
//       
//       c = 0, /* case bonus (0 to the length of max(str1,str2)) */
//       s = 0, /* Start of string bonus (0 or 1) */
//       b = 0, /* beginning of word bonus (0 to 1*(# of words)) */
//       
//       dj = 0,                   /* distance Jaro */
//       w = 0, /* number of words (not a bonus)*/
//       isMatch = function(a,b){
//         return RegExp.new(a,true).test(b);
//       };
// 
//       
//   // Test if swapping strX & lenStrX if stra is longer then str2 for proformance ??
//   // another option is to bail out of the stepping once we are outside of the context of the other string
//   // the issue is that with string lengths of 11 & 2 you wouldn't want to go through the loop 11 times
//   
//   /* find m & t */
//   for (var i in str2) {
//     // letter = str2[i];
//     if(str1.slice(i,i+w).indexOf(str2[i]) > -1) { /* match */
//       m++;
//     } else if(str1.slice(i-w,i).indexOf(str2[i]) > -1) { /* transposition */
//       m++; t++;
//     }
//   };
//   return (1/3*(m/s1+m/s2+(m-t)/m));
//   // dj = (1/3*(m/s1+m/s2+(m-t)/m));
//   
//   
//   
//   
//   // var p = 0.1;  
//   //   var l=0; 
//   //   
//   //   for(var i=0; i<4; i++) { /* find length of prefix match (max 4) */
//   //     if(str1[i]==str2[i]){ l++; } else { break; } 
//   //   }
//   
//   // return dj+((s*s1+c+(b/w)*s1)/(1/(s1*3))(1-dj));
// };
// 
// 
// 
// 
String.prototype.score = function(abbreviation) {
  return jaro(this.toLowerCase(),abbreviation.toLowerCase());
};
// 
// // console.log('martha '+'martha'.score('marhta'));
// // console.log('martha '+jaroWinkler('martha','marhta'));
// // console.log('dwayne '+jaroWinkler('DWAYNE','DUANE'));