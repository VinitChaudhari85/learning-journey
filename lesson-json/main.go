package main

import (
	"encoding/json"
	"fmt"
)

type Gundam struct {
	GundamGrade string `json:"gundam-grade"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
}

func main() {
	goText := Gundam{GundamGrade: "HG", Name: "Schwarzette", Price: 2999}

	//encoding
	data, err := json.Marshal(goText)
	if err != nil {
		fmt.Println("Marshal error!")
	}
	fmt.Println(string(data))

	//decoding
	jsonText := `{"gundam-grade": "HG", "name": "Schwarzette", "price": 2999}`
	var receiving Gundam
	err = json.Unmarshal([]byte(jsonText), &receiving)
	if err != nil {
		fmt.Println("Unmarshal error!")
	}
	fmt.Println(receiving.GundamGrade, receiving.Name, receiving.Price)
}
