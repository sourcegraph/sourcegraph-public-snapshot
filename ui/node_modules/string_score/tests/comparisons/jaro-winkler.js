jaro = function(str1, str2){
  var lenStr1 = str1.length,
      lenStr2 = str2.length,
      matchWindow = Math.max(lenStr1, lenStr2)/2-1,
      transpositions=0,
      matches=0,
      letter='';
      
  // Test if swapping strX & lenStrX if stra is longer then str2 for proformance ??
  // another option is to bail out of the stepping once we are outside of the context of the other string
  // the issue is that with string lengths of 11 & 2 you wouldn't want to go through the loop 11 times
  
  /* find matches & transpositions */
  for (var i in str2) {
    letter = str2[i];
    if(str1.slice(i,i+matchWindow).indexOf(letter) > -1) { /* match */
      matches++;
    } else if(str1.slice(i-matchWindow,i).indexOf(letter) > -1) { /* transposition */
      matches++; transpositions++;
    }
  };
  return (1/3*(matches/lenStr1+matches/lenStr2+(matches-transpositions)/matches));
};

jaroWinkler = function(str1, str2, p){
  p = p || 0.1;
  var dj = jaro(str1,str2), l=0; 
  
  for(var i=0; i<4; i++) { /* find length of prefix match (max 4) */
    if(str1[i]==str2[i]){ l++; } else { break; } 
  }
  
  return dj+(l*p*(1-dj));
};

String.prototype.score = function(abbreviation) {
  return jaroWinkler(this.toLowerCase(),abbreviation.toLowerCase());
};

// console.log('martha '+'martha'.score('marhta'));
// console.log('martha '+jaroWinkler('martha','marhta'));
// console.log('dwayne '+jaroWinkler('DWAYNE','DUANE'));