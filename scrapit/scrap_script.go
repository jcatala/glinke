package scrapit



import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
)

func GrabLinks(client *http.Client, url string, v bool) []string{
	// if the url does not terminate on /, add it
	if url[len(url)-1] != '/' {
		url = url + "/"
	}
	var ret []string
	// logic of grabbing links :D
	res, err := client.Get(url)
	if err != nil {
		if v {
			log.Println("Error: ", err)
		}
		return nil
	}
	defer res.Body.Close()
	// Load the document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil{
		if v{
			log.Println("Error on creating goquery reader: ", err)
		}
		return nil
	}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		dsrc, _ := s.Attr("data-src")
		if src != "" {
			if !strings.HasPrefix(src, "http"){
				src = url + src
			}
			ret = append(ret, src)
		}
		if dsrc != "" {
			if !strings.HasPrefix(dsrc, "http"){
				dsrc = url + src
			}
			ret = append(ret, dsrc)
		}
	})

	return ret
}