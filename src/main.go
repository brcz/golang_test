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
)

var src_string string

type linkStruct struct {
    URL   string `json:"url"`
    Title string `json:"title"`
}

type responseStruct struct {
    Mentions  []string `json:"mentions,omitempty"`
    Emoticons []string `json:"emoticons,omitempty"`
    Links     []linkStruct   `json:"links,omitempty"`
}

type channelSync struct {}

func init() {
    flag.StringVar(&src_string, "msg", "", "message for parsing")
}

func main() {

    //src_string:= "@bob @john (success) such a cool feature;\nhttps://twitter.com/jdorfman/status/430511497475670016 jr http://twitter.com/jdorfman/status/430511+234 here. \nGood morning! (megusta) (coffee)"
    
    flag.Parse()
    
    if result, err := ParseInput(src_string); err!= nil {
        fmt.Println("Error: ", err.Error())
    } else {
        fmt.Println("Result: ", result)
    }

}

func ParseInput(source string) (result string, err error) {
    var bytes []byte
    var parsedStruct responseStruct
    
    mentions := findMentions(source)
    emoticons := findEmoticons(source)
    urls := findLinks(source)
    
    parsedStruct = responseStruct{mentions, emoticons, urls}
    
    bytes, err = json.Marshal(parsedStruct)
    result = string(bytes)
    return
}


func findMentions(source string) (mentions []string){
    mentionsRegexpString:=`\@(?P<tag>\w+)`
    mentions,_ = getMatches(mentionsRegexpString, source)
    
    return
}

func findEmoticons(source string) (emoticons []string){
    emoticonsRegexpString:=`\((?P<tag>\w{1,15})\)`
    emoticons,_ = getMatches(emoticonsRegexpString, source)
    
    return
}

func findLinks(source string) (links []linkStruct){
    links = []linkStruct{}
    
    linkRegexpString:=`(?P<tag>(((https?|ftp|file)://?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|/))))`
    linkMatches,_ :=getMatches(linkRegexpString, source)
    
    if linkMatches !=nil {
        // Channels
        chUrls := make(chan linkStruct)
        chFinished := make(chan channelSync)
        
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

func fetchUrlAndParseTitle(url string, ch chan linkStruct, chFinished chan channelSync) {
    
    defer func() {
        // Notify that we're done after this function
        chFinished <- channelSync{}
    }()
 
    title := ""
    
    response, err := http.Get(url)
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

func extractTitle(r io.Reader) (title string, err error) {
    
    doc, err := html.Parse(r)
    if err != nil {
        log.Println("Fail to parse html")
        return
    }
    
    return traverse(doc)
}

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

