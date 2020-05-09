package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

const omniaBaseURL = "https://b2b.omniacomponents.com/"

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	logIntoOmnia(client)
	getItems(client, 1327)
}

func logIntoOmnia(c *http.Client) {
	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type loginResponse struct {
		User *struct{} `json:"user"`
	}

	res, err := c.Do(omniaJSONRequest("POST", "login", loginRequest{
		Username: "Generic Customer",
		Password: "gen_cust_2019",
	}))

	if err != nil {
		panic(err)
	}

	var response loginResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		panic(err)
	}

	if response.User == nil {
		panic("failed login, the credentials are likely incorrect")
	}
}

func omniaJSONRequest(method, endpoint string, data interface{}) *http.Request {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(data); err != nil {
		panic(err)
	}

	r, err := http.NewRequest(method, omniaBaseURL+endpoint, body)
	if err != nil {
		panic(err)
	}

	addHeaders(r)
	return r
}

func addHeaders(r *http.Request) {
	r.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:76.0) Gecko/20100101 Firefox/76.0")
	r.Header.Add("Content-Type", "application/json")
}

type item struct {
	Code          string   `json:"code"`
	OriginalCodes []string `json:"original_codes"`
	Description   string   `json:"description"`
	ImageURL      string   `json:"image_url"`
}

func getItems(c *http.Client, categoryID int) []item {
	var result []item

	for i := 1; ; i++ {
		productList := getProductList(c, categoryID, i)

		if len(productList.Products) == 0 {
			return result
		}

		for _, product := range productList.Products {
			randomDelay()

			technicalDetails := getTechnicalDetails(c, product.ID)
			item := item{
				Code:          product.Code,
				Description:   product.Name + *technicalDetails.TechnicalDescription,
				ImageURL:      product.Image,
				OriginalCodes: strings.Split(*technicalDetails.OriginalCodes, ","),
			}
			result = append(result, item)
		}
	}
}

func randomDelay() {
	n := time.Duration(rand.Int63n(1000) + 100)
	time.Sleep(n * time.Millisecond)
}

type productListRequest struct {
	CategoryID     int       `json:"category_id"`
	DivisionID     string    `json:"division_id"`
	OnlyAvailable  *struct{} `json:"onlyAvailable"` // Always nil
	OrderBy        string    `json:"orderBy"`
	PageIndex      int       `json:"page_index"`
	PageSize       int       `json:"page_size"`
	SelectedFacets string    `json:"selected_facets"`
	UserSearch     string    `json:"user_search"`
}

type productListResponse struct {
	Products []product `json:"products"`
}

type product struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func getProductList(c *http.Client, categoryID int, index int) *productListResponse {
	res, err := c.Do(omniaJSONRequest("POST", "api/v1/public/get_productlist", &productListRequest{
		CategoryID:     categoryID,
		DivisionID:     "1",
		OnlyAvailable:  nil,
		OrderBy:        "price asc",
		PageIndex:      index,
		PageSize:       20,
		SelectedFacets: "",
		UserSearch:     "",
	}))

	if err != nil {
		panic(err)
	}

	productListResponse := &productListResponse{}
	if err := json.NewDecoder(res.Body).Decode(&productListResponse); err != nil {
		panic(err)
	}
	return productListResponse
}

type techsheetRequest struct {
	ProductID string     `json:"product_id"`
	Filter    []struct{} `json:"filter"`
}

type technicalDetails struct {
	OriginalCodes        *string `json:"cross_reference_customer"`
	TechnicalDescription *string `json:"technical_description"`
}

type techsheetData struct {
	General []technicalDetails `json:"dati_generali"`
}

type techsheetResponse struct {
	Data techsheetData `json:"data"`
}

func getTechnicalDetails(c *http.Client, productID int) *technicalDetails {
	res, err := c.Do(omniaJSONRequest("POST", "api/v1/public/get_techsheet_data", techsheetRequest{
		ProductID: strconv.Itoa(productID),
		Filter:    []struct{}{},
	}))
	if err != nil {
		panic(err)
	}
	var techsheetRes []techsheetResponse
	if err := json.NewDecoder(res.Body).Decode(&techsheetRes); err != nil {
		panic(err)
	}

	return &techsheetRes[0].Data.General[0]
}
