package utils

import (
	"errors"
	"reflect"
	"time"
	"encoding/json"
	"io"
	"net/http"
)

func Time(strct interface{}, new bool) error {
	t := reflect.TypeOf(strct)
	v := reflect.ValueOf(strct).Elem()
	// validate if strct is a pointer and struct
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("strct must be a pointer to a struct")
	}

	// validate if the datecreated and dateupdated fields are in strct and are of type time.Time
	dateCreatedField, has_created := t.Elem().FieldByName("DateCreated")
	dateUpdatedField, has_updated := t.Elem().FieldByName("DateUpdated")
	if has_created == false || has_updated == false {
		return errors.New("strct must have DateCreated and DateUpdated fields")
	}

	if dateCreatedField.Type.Kind() != reflect.TypeOf(time.Time{}).Kind() || dateUpdatedField.Type.Kind() != reflect.TypeOf(time.Time{}).Kind() {
		return errors.New("strct DateCreated and DateUpdated fields must be of type time.Time")
	}

	// set the time for the fields based on new arguement value
	if new {
		// Set the "DateUpdated" field to current UTC time
		v.FieldByName("DateCreated").Set(reflect.ValueOf(time.Now().UTC()))
	}
	v.FieldByName("DateUpdated").Set(reflect.ValueOf(time.Now().UTC()))
	return nil

}



// sends json response
func JsonResponse(w http.ResponseWriter, success bool, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success:  success,
		Message: message,
		Data:    data,
	})
}

// unmarshall io.Reader type such as http request body
func Unmarshall(r io.Reader, strct interface{})(error){
	structType := reflect.TypeOf(strct)
	if structType.Kind() != reflect.Ptr || structType.Elem().Kind() != reflect.Struct {
		return errors.New("strct must be pointer to a struct")
	}

	err:=json.NewDecoder(r).Decode(strct)
	if err != nil{
		return err
	}
	return nil
}