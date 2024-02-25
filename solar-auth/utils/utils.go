package utils

import (
	"errors"
	"reflect"
	"time"
	"encoding/json"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"net/http"
	"crypto/ed25519"
	"crypto/rand"
	"strconv"
	"strings"
	mathrand "math/rand"
	"github.com/golang-jwt/jwt/v5"
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

// Generate user access key
func GenerateAccessKey(email string) (string,error){
	var (
		accessKey string
		err error
		signingKey ed25519.PrivateKey //since PrivateKey implements crypto.Signer
	)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, UserClaim{
		RegisteredClaims: jwt.RegisteredClaims{},
		Email: email,
		DateTime: time.Now().UTC(),
	})

	// Create the actual key
	signingKey, err = hex.DecodeString(os.Getenv("SIGNING_KEY"))
	if err != nil {
		return "", err
	}

	accessKey, err = token.SignedString(signingKey)

	if err != nil {
		return "", err
	}
	return accessKey, nil
}

func GenerateKeyPairs()([]byte, []byte, error){
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil{
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}

func GenerateUserID(email string)string {
	// Modify email address (example: remove special characters)
	var (
		sanitized strings.Builder
		idLength int = 24
	)
	for _, ch := range email {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			sanitized.WriteRune(ch)
		}
	}
	emailSanitized := sanitized.String()

	// Get current time in milliseconds
	now := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate a random integer within a specific range using crypto/rand for better randomness
	randomIntBytes := make([]byte, 8) // 8 bytes for int64
	rand.Read(randomIntBytes)
	randomInt := int64(binary.LittleEndian.Uint64(randomIntBytes)) % now
	
	// Combine modified email, timestamp, and random integer
	combined := emailSanitized + strconv.FormatInt(now, 10) + strconv.FormatInt(randomInt, 10)

	// Ensure the combined string is exactly 24 characters long
	if len(combined) > idLength {
		combined = combined[:idLength]
	} else if len(combined) < idLength {
		combined += strings.Repeat("-", idLength-len(combined))
	}

	// Shuffle the combined string
	runes := []rune(combined)
	for i := len(runes) - 1; i > 0; i-- {
		j := mathrand.Intn(i + 1) // Generate a random index within the remaining range
		runes[i], runes[j] = runes[j], runes[i] // Swap characters at indices i and j
	}
	return string(runes)
}


