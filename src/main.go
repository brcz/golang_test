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
)

type linkStruct struct {
    URL   string `json:"url"`
    Title string `json:"title"`
}

type responseStruct struct {
    Mentions  []string `json:"mentions,omitempty"`
    Emoticons []string `json:"emoticons,omitempty"`
    Links     []linkStruct   `json:"links,omitempty"`
}

func main() {

    src_string:= "@bob @john (success) such a cool feature;\nhttps://twitter.com/jdorfman/status/430511497475670016 jr http://twitter.com/jdorfman/status/430511+234 here. \nGood morning! (megusta) (coffee)"
    
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
    mentionsRegexpString:=`\@\w+`
    mentionsRegexp,_ := regexp.Compile(mentionsRegexpString)
    mentionsMatches := mentionsRegexp.FindAllString(source, -1)
    
    if mentionsMatches !=nil {
        for i, mention := range mentionsMatches {
            mentionsMatches[i] = mention[1:len(mention)]
        }
    }
    
    return mentionsMatches
}

func findEmoticons(source string) (emoticons []string){
    emoticonsRegexpString:=`\(\w+\)`
    emoticonsRegexp,_ := regexp.Compile(emoticonsRegexpString)
    emoticonsMatches := emoticonsRegexp.FindAllString(source, -1)
    
    if emoticonsMatches !=nil {
        for i, emoticon := range emoticonsMatches {
            emoticonsMatches[i] = emoticon[1:len(emoticon)-1]
        }
    }
    
    return emoticonsMatches
}

func findLinks(source string) (links []linkStruct){
    links = []linkStruct{}
    
    linkRegexpString:=`((https?://?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|/)))`
    linkRegexp,_ := regexp.Compile(linkRegexpString)
    linkMatches := linkRegexp.FindAllString(source, -1)
    
    if linkMatches !=nil {
        // Channels
        chUrls := make(chan linkStruct)
        chFinished := make(chan bool)
        
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

func fetchUrlAndParseTitle(url string, ch chan linkStruct, chFinished chan bool) {
    
    defer func() {
        // Notify that we're done after this function
        chFinished <- true
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

