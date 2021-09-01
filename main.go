package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gorilla/mux"
)

// URL is struct of body
type URL struct {
	URL string `json:"url"`
}

// Body is struct of body
type Body struct {
	NAME         string `json:"name"`
	IMAGEURL     string `json:"imageURL"`
	DESCRIPTION  string `json:"description"`
	PRICE        string `json:"price"`
	TOTALREVIEWS string `json:"totalreviews"`
	URL          string `json:"url"`
}

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the HomePage!")
	var body URL
	var dbpayload []Body
	var tempDbpayload Body

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return
	}

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnHTML("div.s-result-list.s-search-results.sg-row", func(e *colly.HTMLElement) {
		e.ForEach("div.a-section.a-spacing-medium", func(_ int, e *colly.HTMLElement) {

			productName := e.ChildText("span.a-size-medium.a-color-base.a-text-normal")

			if productName == "" {
				// If we can't get any name, we return and go directly to the next element
				return
			}

			price := e.ChildText("span.a-price > span.a-offscreen")
			FormatPrice(&price)

			review := e.ChildText("span.a-size-base")
			// fmt.Printf("Product Name: %s \nPrice: %s \n Review count: %s \n", productName, price, review)

			tempDbpayload.DESCRIPTION = productName
			tempDbpayload.NAME = productName
			tempDbpayload.IMAGEURL = ""
			tempDbpayload.PRICE = price
			tempDbpayload.TOTALREVIEWS = review
			tempDbpayload.URL = body.URL
			dbpayload = append(dbpayload, tempDbpayload)
		})
	})

	c.Visit(body.URL)
	fmt.Println("Calling API...")
	dbpayload = append(dbpayload, tempDbpayload)
	dataByte, err := json.Marshal(dbpayload)
	if err != nil {
		return
	}
	bodyString := strings.NewReader(string(dataByte))
	URL := "http://host.docker.internal:10091"
	fmt.Printf("url for Post %v \n", URL)
	resp, statusCode := processRequest(bodyString, URL, "POST")
	fmt.Println(statusCode)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("not able to read a response from sellerdb POST API %v \n", err)
		return
	}
	fmt.Printf("Response Body from post sellerdb API %v \n", string(responseBody))

}

func handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter()
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/", homepage).Methods("POST")
	log.Fatal(http.ListenAndServe(":10090", myRouter))
}
func main() {
	handleRequests()
}

// FormatPrice fetch the Price
func FormatPrice(price *string) {
	r := regexp.MustCompile(`\$(\d+(\.\d+)?).*$`)

	newPrices := r.FindStringSubmatch(*price)

	if len(newPrices) > 1 {
		*price = newPrices[1]
	} else {
		*price = "Unknown"
	}
}

func processRequest(body *strings.Reader, URL string, method string) (response *http.Response, statusCode int) {
	tr := &http.Transport{
		// This is the insecure setting, it should be set to false.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	var req *http.Request
	var err error
	if method == "POST" {
		req, err = http.NewRequest(method, URL, body)
		if err != nil {
			fmt.Println("not able  to make request", err)
			return
		}
	}

	if err != nil {
		fmt.Printf("not able to create a request for seller db %v \n", err)
		return nil, 400
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("return error from seller db %v \n", err)
		return nil, resp.StatusCode
	}
	return resp, resp.StatusCode

}
