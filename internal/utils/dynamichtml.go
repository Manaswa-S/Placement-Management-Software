package utils

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
)

func DynamicHTML(pathToHTML string, data interface{}) (bytes.Buffer, error) {

	// embed the token in the html of the pass post page
	var body bytes.Buffer
	// Parse the template file into object assigned to 'bodytemp'
	bodytemplate, err := template.ParseFiles(pathToHTML)
	if err != nil {
		fmt.Println(err.Error())
		return bytes.Buffer{}, errors.New("failed to parse html template")
	}

	// Execute the template and apply 'data' to the template
	// store the formed result in 'body'
	err = bodytemplate.Execute(&body, data)
	if err != nil {
		fmt.Println(err.Error())
		return bytes.Buffer{}, errors.New("failed to execute html template")
	}

	return body, nil
}