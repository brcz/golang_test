package main

import (
        "encoding/json"
        "fmt"
        "regexp"
        "net/http"
        "golang.org/x/net/html"
        "io"
        "errors"
        "log"
        "flag"
		  "time"
)

// source string with message to be parsed
var src_string string

// struct to gather link url, title and create json object
type linkStruct struct {
    URL   string `json:"url"`
    Title string `json:"title"`
}

// struct to create final json object
type responseStruct struct {
    Mentions  []string `json:"mentions,omitempty"`
    Emoticons []string `json:"emoticons,omitempty"`
    Links     []linkStruct   `json:"links,omitempty"`
}

// zero-sized type to sync with title gathering goroutine
type channelSync struct {}

//comment this

func init() {
	//bind msg flag to src_string variable
    flag.StringVar(&src_string, "msg", "", "message for parsing")
}

func main() {

	 flag.Parse() //retrieve command line arguments

	// try to parse recieved string and print json string or error message
    if result, err := ParseInput(src_string); err!= nil {
        fmt.Println("Error: ", err.Error())
    } else {
        fmt.Println("Result: ", result)
    }

}

/***
* parsing a message string to find mentions, emoticons and URLS.
*
* input: message string
* output: JSON-formatted string
*
*/
func ParseInput(source string) (result string, err error) {
    var bytes []byte
    var parsedStruct responseStruct

	// can be paralleled
    mentions := findMentions(source)
    emoticons := findEmoticons(source)
    urls := findLinks(source)

    parsedStruct = responseStruct{mentions, emoticons, urls}

    bytes, err = json.Marshal(parsedStruct)
    result = string(bytes)
    return
}

/***
* function to find mentions in a string.
*
* input: message string
* output: slice of strings with found mentions without @ sign
*
*/
func findMentions(source string) (mentions []string){
    mentionsRegexpString:=`\@(?P<tag>\w+)`
    mentions,_ = getMatches(mentionsRegexpString, source)

    return
}

/***
* function to find emoticons in a string.
*
* input: message string
* output: slice of strings with found emoticons without framing parenthesis
*
*/
func findEmoticons(source string) (emoticons []string){
    emoticonsRegexpString:=`\((?P<tag>\w{1,15})\)`
    emoticons,_ = getMatches(emoticonsRegexpString, source)

    return
}

/***
* function to find URL and retrieve Title for it in a string.
*
* input: message string
* output: slice of linkStruct struct containing finded URL and title
*
*/
func findLinks(source string) (links []linkStruct){
    links = []linkStruct{}

    linkRegexpString:=`(?P<tag>(((https?|ftp|file)://?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|/))))`
    linkMatches,_ :=getMatches(linkRegexpString, source)

    if linkMatches !=nil {
        // Channels for communacation
        chUrls := make(chan linkStruct) //fetch titles for parsed URLS
        chFinished := make(chan channelSync) //finish sync

        for _, url := range linkMatches {
            go fetchUrlAndParseTitle(url, chUrls, chFinished)
        }

        for c := 0; c < len(linkMatches); {
            select {
                case titledURL := <-chUrls:
                    links = append(links, titledURL)
                case <-chFinished:
                    c++
            }
        }

    }

    return
}

/***
* function to process given RegExp and extract <tag> named part.
* using named parts scince Go's "regxep" doesn't support PCRE
*
* input: regexp expression string
* output: slice of strings containing finded named parts
*
*/
func getMatches(expr string, source string) (matches []string, err error){
    pattern := regexp.MustCompile(expr)
    mtchs:= pattern.FindAllStringSubmatch(source, -1)
    index := 0
    for i, name := range pattern.SubexpNames() {
            if name == "tag" {
                index = i
                break
            }
    }
    for _, submatch := range mtchs {
        matches = append(matches, submatch[index])
    }
    return
}

/***
* function to fetch data from given URL and extract value of <title> tag.
*
* input: URL string, 2 channels to retrieve data and "function finished" signal
* output: linkStruct for given URL and channelSync finished signal
*
*/
func fetchUrlAndParseTitle(url string, ch chan linkStruct, chFinished chan channelSync) {

    defer func() {
        // Notify that we're done after this function
        chFinished <- channelSync{}
    }()

    title := ""
	//fmt.Println("url=" ,url)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
			}
    response, err :=netClient.Get(url)
    if err != nil {
        log.Println("ERROR: Failed to fetch \"" + url + "\"")
        //return
    }

    defer response.Body.Close()

    title, err = extractTitle(response.Body);

    if err!=nil {
        log.Println("ERROR: Failed to get HTML title")
    }

    ch <- linkStruct{url, title}

    return
}

/***
* function to process title extraction from
*
* input: regexp expression string
* output: finded title string
*
*/
func extractTitle(r io.Reader) (title string, err error) {

    doc, err := html.Parse(r)
    if err != nil {
        log.Println("Fail to parse html")
        return
    }

    return traverse(doc)
}

/***
* recursive function to pass on html markup and extract value of <title> tag
*
* input: pointer on current html node
* output: value of <title> tag or empty string if nothing found.
*
*/
func traverse(n *html.Node) (string, error) {
    if n.Type == html.ElementNode && n.Data == "title" {
        return n.FirstChild.Data, nil
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        result, err := traverse(c)
        if err==nil {
            return result, err
        }
    }

    return "", errors.New("false")
}
