// Algorithm: Find all the possible ways the abbr could match the string, then using the various weighting options, return the greatest score of all of them.

// Tests:
//  + A string that matches exactly, including the case, has a similar rating of 1.0
//  + A string that matches exactly, except for the case, has a similar rating of 0.9
//  + A string that has all characters in common, but none are in the same order, has a similar rating of 0.4
//  + A string that has zero common characters has a similar rating of 0.0

// To create these scores, each character must be given a weight to its own score, based on the proportive size of the given and test strings.
// The given string must be given a base score as well, based on how many words, capital letters, etc, it has.
// Each bonus is given to a character based on the total potential score of that bonus in the given string.

// All of these are score MODIFIERS - each test character starts with a base score of 1, each applicable boost adds to (or subtracts from) its overall score.
String.scoring_options = {
  // just adds on to the acronym boost
  firstChar : 0.5,
  // such as 'qbf' or 'QBF' when referring to 'The Quick Brown Fox'
  acronym : 2.2,
  // match boost for matching capital letters, whether query was capital or not. Thus, 'ABC'.score('ABC') > 'abc'.score('abc')
  capitalLetter : 0.8,
  // when the query character matches case-sensitive
  caseMatch : 0.3,
  // when two consecutive characters of the query match two consecutive characters of the string
  consecutiveChars : 0.5,
  nonConsecutiveChars : -0.2,
  // when a query character is missing from the string
  missingMatch : -6,
  // when a query character matches in the string, but is not in order with the rest of the string
  outOfOrder : {
    // Subtracts this proportionally based on the ACTUAL point value of each character in the abbreviation,
    // and then reduces the other boosts by the multiplier value. That way we're SURE it's always positively
    // valuable to include a character even if it's out of place.
    score: -1.5, // subtracts all of the base value of having the character in there - only bonuses count now.
    multiplier: 0.1 // reduces the effect of the other boosts by a LOT - although if there are any boosts at all, the score will at least be positive
  }
};

Array.prototype.remove = function(index){
  return this.splice(index,1)[0];
};
Array.prototype.includes = function(value){
  var i,len=this.length;
  for(i=0;i<len;i++){
    if(this[i]===value)return true;
  }
  return false;
};
Array.prototype.count_how = function(how){
  var i,count=0,len=this.length;
  for(i=0;i<len;i++){
    if(how.apply(this[i]))count+=1;
  }
  return count;
};
Array.prototype.each = function(cb){
  var i,len=this.length;
  for(i=0;i<len;i++){
    cb.apply(this[i], [i]);
  }
  return this;
};
Array.prototype.map = function(cb){
  var i,len=this.length,a=[];
  for(i=0;i<len;i++){
    a.push(cb.apply(this[i]));
  }
  return a;
};
Array.prototype.highest = function(how){
  return this[this.highest_i(how)];
};
Array.prototype.highest_i = function(how){
  var i,s,best_s=this[0],best_i=0,len=this.length;
  for(i=0;i<len;i++){
    if(how)s=how(this[i]);
    else s=this[i];
    if(s>best_s){
      best_s=s;
      best_i=i;
    }
  }
  return best_i;
};
String.prototype.first = function(){
  return this.slice(0,1);
};
String.prototype.count_match = function(regexp){
  var str=''+this,count=0;
  var pos=str.search(regexp);
  while(pos>-1){
    count += 1;
    str=str.slice(pos+1);
    pos=str.search(regexp);
  }
  return count;
};

var MatchTree = function(parent, orig_string, string, abbr, positions){
  this.ancestry = function(){
    return this.parent ? this.parent.ancestry().concat([this]) : [];
  };

  this.paths = function(){
    var paths=[],abbr_so_far;
    if(this.next_matches){
      var i,mlen=this.next_matches.length;
      for(i=0;i<mlen;i++){
        paths = paths.concat(this.next_matches[i].paths());
      }
    }else{
      paths.push(this.ancestry());
    }
    return paths;
  };

  // Constructor
  if(parent.constructor===String){
    this.parent = null;
    this.original = ''+parent;

    this.positions = [];
    this.position = 0;

    abbr = orig_string;
    string = parent;
    orig_string = string;
  }else{
    this.parent = parent;
    this.original = parent.original; // pass the string down each level
    this.positions = positions;
    this.position = parent.original.length - string.length;

    // In this case, the abbr that is sent matches on the first character. Here we analyze that match before matching further.
    var match_chr = string.first();
    this.chr = abbr.first();

    // Add bonuses
    this.match_info = [this.chr + this.position];
    if(match_chr.toLowerCase()!==this.chr.toLowerCase()){
      this.match_info.push('missingMatch');
      this.position=parent.position;
    }else{
      if(this.position<this.parent.position)this.match_info.push('outOfOrder');
      if(this.position===0)this.match_info.push('firstChar');
      if(this.position===0 || this.original.slice(this.position-1,this.position)==' ')this.match_info.push('acronym');
      if(match_chr.toUpperCase()===match_chr)this.match_info.push('capitalLetter');
      if(this.chr===match_chr)this.match_info.push('caseMatch');
      if(this.position!==0 && this.parent.position===this.position-1)this.match_info.push('consecutiveChars');
        else this.match_info.push('nonConsecutiveChars');
    }
    
    // Shift over to the next abbreviation character
    string=''+this.original;
    abbr = abbr.slice(1);
  }

  // If there is any more abbr to match, then create children
  if(abbr.length>0){
    this.next_matches = [];
    // Create a new MatchTree for each possible match, and save them in this.matches
    var chr=abbr[0],pos=string.toLowerCase().indexOf(chr.toLowerCase()),pp;
    while(pos>-1){
      pp=this.original.length-string.length+pos;
      // Can't match the same character twice.
      if(!this.positions.includes(pp))
        this.next_matches.push(new MatchTree(this, orig_string, string.slice(pos), abbr, this.positions.concat([pp])));
      string=string.slice(pos+1);
      pos=string.toLowerCase().indexOf(chr.toLowerCase());
    }
    if(this.next_matches.length===0){
      this.next_matches.push(new MatchTree(this, orig_string, string, abbr, this.positions));
    }
  }
};

var CachedScores = {};

String.prototype.score = function(abbr){
  // Use cached version if we've already scored this string for this abbr!
  if(CachedScores[this] && CachedScores[this][abbr]) return CachedScores[this][abbr];
  // Cheat: if exact match, go ahead and just immediately return the highest score possible
  if(this==abbr)return 1.0;
  
  // Set up the scoring options. These are all adding to or subtracting from a regular fact-of-match = 1
    var options = String.scoring_options || {};
    // when the first character of the string matches the first character of the query
    if(!options.firstChar) options.firstChar = 0.5; // just adds on to the acronym boost
    // such as 'qbf' or 'QBF' when referring to 'The Quick Brown Fox'
    if(!options.acronym) options.acronym = 1;
    // match boost for matching capital letters, whether query was capital or not. Thus, 'ABC'.score('ABC') > 'abc'.score('abc')
    if(!options.capitalLetter) options.capitalLetter = 0.2;
    // when the query character matches case-sensitive
    if(!options.caseMatch) options.caseMatch = 0.2;
    // when two consecutive characters of the query match two consecutive characters of the string
    if(!options.consecutiveChars) options.consecutiveChars = 0.2;
    // when a query character is missing from the string
    if(!options.missingMatch) options.missingMatch = -5;
    // when a query character matches in the string, but is not in order with the rest of the string
    if(!options.outOfOrder) options.outOfOrder = {
      // Subtracts this proportionally based on the ACTUAL point value of each character in the abbreviation,
      // and then reduces the other boosts by the multiplier value. That way we're SURE it's always positively
      // valuable to include a character even if it's out of place.
      score: -0.95, // subtracts most of the value of having the character in there
      multiplier: 0.1 // reduces the effect of the other boosts by quite a bit - although if there are several boosts, the score will still amount to at least something
    };

    // Other ideas for weights:
    //   +bump related to the number of matching characters there are in the string that follow the last match
    //   +bump for how many characters in the string match the test character
    //   +bump for close proximity
    //   +bump related to the size of the abbreviation compared to the size of the test string
    //   +bump for a higher number of matching consecutive characters

    // Find all possible match paths
    var match_tree = new MatchTree(this, abbr);

    // Secondly, determine a potential score on the base string for each bonus type.
    var potential_scores = {
      // how many /\s\w/ in the string (include the first character)?
      words: this.count_match(/\s\w/)+1,
      // how many capital letters in the string?
      capitals: this.count_match(/[A-Z]/),
      // how many characters in the string
      length: this.length
    };
    var potential_score =
      options.firstChar +
      potential_scores.length +
      potential_scores.length * (0.0 + options.caseMatch) +
      (potential_scores.length-1) * (0.0 + options.consecutiveChars) +
      potential_scores.words * (0.0 + options.acronym) +
      potential_scores.capitals * (0.0 + options.capitalLetter);
    
    // Thirdly, give scores to each matched character, proportional to the potential score for each bonus type.
    // proportion of the square roots of the lengths?
    // var proportion = Math.sqrt(abbr.length) / Math.sqrt(this.length);
    var proportion = abbr.length / this.length;
    var score_per_character = potential_score * proportion / abbr.length;

    // Now, score each match path
    var paths = match_tree.paths();
    var path,plen=paths.length;

    // Then, calculate the score for each path
    var i,j,match_infos,score,scores=[],multiplier,_char;
    paths.each(function(){
      path = this;
      score = 0;
      path.match_infos=[];
      path.each(function(){
        _char = this;
        score += 1;
        multiplier = 1;
        if(this.match_info.includes('outOfOrder'))
          multiplier = options.outOfOrder.multiplier;
        this.match_info.each(function(){
          if(options[this])
            score += (
              this=='outOfOrder' ? (options.outOfOrder.score * (score_per_character-1)) :
              // This score gets worse depending on the proximity of the two characters' match locations
              ( this=='nonConsecutiveChars' ? (((_char.position - _char.parent.position) / abbr.length) * (options.nonConsecutiveChars * 2)) :
                (options[this] * multiplier)
              )
            );
        });
        path.match_infos.push(this.match_info, score);
      });
      // Extra: first-abbr-character-is-on-acronym boost - 2/3 of firstChar boost
      if(path[0].match_info.includes('acronym') && !path[0].match_info.includes('firstChar'))
        score += (options.firstChar * 3/4);
      scores.push(score);
    });

    // Last, return the proportion of the best score to the base potential score.
    var best_path_i = scores.highest_i();
    var best_path = paths[best_path_i];

    var highest = scores[best_path_i];
    
    var final_score = highest===0 ? 0 : highest/potential_score;

    // Cache the resulting score!
    if(!CachedScores[this]) CachedScores[this] = {};
    CachedScores[this][abbr] = final_score;

    return final_score;
};

Array.prototype.best_score_index = function(abbr){
  return this.highest_i(function(e){
    return e.score(abbr);
  });
};
