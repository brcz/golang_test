package main

import (
        "testing"
)

//types for testing datasets
type testPair struct {
    sampleSource string
    expectedResult string
}

type testMentionsPair struct {
		sampleSource string
		expectedResult []string
}

type testEmoticonsPair struct {
		sampleSource string
		expectedResult []string
}

type testLinksPair struct {
		sampleSource string
		expectedResult []linkStruct
}

// test dataset for ParseInput. If you want to expand test dataset use separate tests instead.
var testJSON = []testPair{
    {   "@chris you around?",
        `{"mentions":["chris"]}` },

    {   "Good morning! (megusta) (coffee)",
        `{"emoticons":["megusta","coffee"]}` },

    {   "Olympics are starting soon; http://www.nbcolympics.com",
        `{"links":[{"url":"http://www.nbcolympics.com","title":"2016 Rio Olympic Games | NBC Olympics"}]}` },

    {   "@bob @john (success) such a cool feature; \nhttps://twitter.com/jdorfman/status/430511497475670016",
        `{"mentions":["bob","john"],"emoticons":["success"],"links":[{"url":"https://twitter.com/jdorfman/status/430511497475670016","title":"Justin Dorfman on Twitter: \"nice @littlebigdetail from @HipChat (shows hex colors when pasted in chat). http://t.co/7cI6Gjy5pq\""}]}` },

    { "Mentions: @bob fail@bob word @bob  @ bob  ",
			`{"mentions":["bob","bob","bob"]}` },

	 { "Emoticons  (emoticon) (emoticonlongerthanthemaxlimit) (nested (emo)) (nonclosed",
			`{"emoticons":["emoticon","emo"]}` },

	{ "Links: (https://google.com)",
			`{"links":[{"url":"https://google.com","title":"Google"}]}`},
}

// test dataset for mentions
var testMentions = []testMentionsPair {
	{ "@chris you around?", []string{"chris"}},
	{ "Mentions: @bob fail@bob word @bob  @ bob  @joe", []string{"bob","bob","bob","joe"} },
	{ "@bob @john (success) such a cool feature; \nhttps://twitter.com/jdorfman/status/430511497475670016",  []string{"bob","john"} },
}
// test dataset for emoticons
var testEmoticons = []testEmoticonsPair {
	{ "Good morning! (megusta) (coffee)", []string{"megusta", "coffee"} },
	{ "@bob @john (success) such a cool feature; \nhttps://twitter.com/jdorfman/status/430511497475670016", []string{"success"} },
	{ "Emoticons  (emoticon) (emoticonlongerthanthemaxlimit) (nested (emo)) (nonclosed", []string{"emoticon", "emo"} },
}
// test dataset for linnks
var testLinks = []testLinksPair {
	{ "Olympics are starting soon; http://www.nbcolympics.com",
	  []linkStruct{ {"http://www.nbcolympics.com", "2016 Rio Olympic Games | NBC Olympics"} }},
	{ "@bob @john (success) such a cool feature; \nhttps://twitter.com/jdorfman/status/430511497475670016",
	  []linkStruct{ {"https://twitter.com/jdorfman/status/430511497475670016",
						"Justin Dorfman on Twitter: \"nice @littlebigdetail from @HipChat (shows hex colors when pasted in chat). http://t.co/7cI6Gjy5pq\""}}},
	{ "http://google.com, (https://atlassian.com), http://www.fruitywifi.com/index_eng.html",
	  []linkStruct{ {"http://google.com", "Google"},
					    {"https://atlassian.com","Software Development and Collaboration Tools | Atlassian"},
						 {"http://www.fruitywifi.com/index_eng.html","FruityWifi"}} },
}

func TestParseInput(t *testing.T) {
    for _, pair := range testJSON {
        result,_ := ParseInput(pair.sampleSource)

        if result != pair.expectedResult {
            t.Error(
                    "For", pair.sampleSource,
                    "\nexpected", pair.expectedResult,
                    "\ngot", result,
                    )
        }
    }
}

func TestFindMentions(t *testing.T) {
	for _, pair := range testMentions {
        result := findMentions(pair.sampleSource)
			//first check for sufficiency
			for _,value := range result {
				if found, _ := searchInStringSlice(value, pair.expectedResult); !found {
					t.Error(
                    "(findMentions)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
					return
				}
			}
			//second check for entirety
			for _,value := range pair.expectedResult {
				if found, _ := searchInStringSlice(value, result ); !found {
					t.Error(
                    "(findMentions)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
				}
			}
    }
}

func TestFindEmoticons(t *testing.T) {
	for _, pair := range testEmoticons {
        result := findEmoticons(pair.sampleSource)
			//first check for sufficiency
			for _,value := range result {
				if found, _ := searchInStringSlice(value, pair.expectedResult); !found {
					t.Error(
                    "(findEmoticons)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
						  return
				}
			}
			//second check for entirety
			for _,value := range pair.expectedResult {
				if found, _ := searchInStringSlice(value, result ); !found {
					t.Error(
                    "(findEmoticons)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
				}
			}
    }
}

func TestFindLinks(t *testing.T) {
	for _, pair := range testLinks {
        result := findLinks(pair.sampleSource)

			//first test for sufficiency
			for _,value := range result {
				if found, _ := searchInLinksSlice(value, pair.expectedResult); !found {
					t.Error(
                    "(findLinks)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
						  return
				}
			}
			//second test for entirety
			for _,value := range pair.expectedResult {
				if found, _ := searchInLinksSlice(value, result ); !found {
					t.Error(
                    "(findLinks)For", pair.sampleSource,
                    "\n expected", pair.expectedResult,
                    "\n      got", result,
                    )
				}
			}
    }
}

// helper function to find string value in string slice
func searchInStringSlice( needle string, haystack []string) (bool, error) {

		for _, val := range haystack {
			if val == needle {
				return true, nil
			}
		}
		return false, nil
}

// helper function to find string value in string slice
func searchInLinksSlice( needle linkStruct, haystack []linkStruct) (bool, error) {

		for _, val := range haystack {
			if val.URL == needle.URL && val.Title == needle.Title {
				return true, nil
			}
		}
		return false, nil
}
