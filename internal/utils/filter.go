package utils

import (
	"errors"
	"reflect"
)

func FilterFromSliceOf[T any](data []T, attribute string, value any) ([]T, error) {

	var result []T
	for _, elem := range data {

		v := reflect.ValueOf(elem)
		if v.Kind() != reflect.Struct {
			return nil, errors.New("data is not a struct")
		}

		field := v.FieldByName(attribute)
		if !field.IsValid() {
			return nil, errors.New("attribute is not valid")
		}

		if reflect.DeepEqual(field.Interface(), value) {
			result = append(result, elem)
		}
	}
	return result, nil
}

func ExtractAfromB[T any](data []T, attribute string) ([]string, error) {

	var result []string
	for _, elem := range data {

		v := reflect.ValueOf(elem)
		if v.Kind() != reflect.Struct {
			return nil, errors.New("data is not a struct")
		}

		field := v.FieldByName(attribute)
		if !field.IsValid() {
			return nil, errors.New("attribute is not valid")
		}

		result = append(result, field.String())
	}
	return result, nil
}