package main

import (
        "testing"
)

type testPair struct {
    sampleSource string
    expectedResult string
}

var tests = []testPair{
    {   "@chris you around?",
        `{"mentions":["chris"]}` },
    
    {   "Good morning! (megusta) (coffee)",
        `{"emoticons":["megusta","coffee"]}` },
    
    {   "Olympics are starting soon; http://www.nbcolympics.com",
        `{"links":[{"url":"http://www.nbcolympics.com","title":"2016 Rio Olympic Games | NBC Olympics"}]}` },
    
    {   "@bob @john (success) such a cool feature; \nhttps://twitter.com/jdorfman/status/430511497475670016",
        `{"mentions":["bob","john"],"emoticons":["success"],"links":[{"url":"https://twitter.com/jdorfman/status/430511497475670016","title":"Justin Dorfman on Twitter: \"nice @littlebigdetail from @HipChat (shows hex colors when pasted in chat). http://t.co/7cI6Gjy5pq\""}]}` },
    { "Mentions: @bob fail@bob word @bob  @ bob  Emoticons  (emoticon) (emoticonlongerthanthemaxlimit) (nested (emo)) (nonclosed", `{"mentions":["bob","bob","bob"],"emoticons":["emoticon","emo"]}` },
}

func TestParseInput(t *testing.T) {
    for _, pair := range tests {
        result,_ := ParseInput(pair.sampleSource)
        
        if result != pair.expectedResult {
            t.Error(
                    "For", pair.sampleSource,
                    "expected", pair.expectedResult,
                    "got", result,
                    )
        }
    }
    
}